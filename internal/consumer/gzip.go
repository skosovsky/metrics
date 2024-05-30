package consumer

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"metrics/internal/log"
)

const methodCompressGzip = "gzip"

type gzipResponseWriter struct {
	http.ResponseWriter
	*responseData
}

func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	if g.statusCode == 0 {
		g.WriteHeader(http.StatusOK)
	}

	g.body = append(g.body, data...)
	g.bodySize += len(data)

	return len(data), nil
}

func (g *gzipResponseWriter) WriteHeader(statusCode int) {
	if g.statusCode == 0 {
		g.statusCode = statusCode
		g.ResponseWriter.WriteHeader(statusCode)
	}
}

func (g *gzipResponseWriter) Header() http.Header {
	return g.ResponseWriter.Header()
}

func WithGzipCompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldDecompress(r, methodCompressGzip) { //nolint:contextcheck // no ctx
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				log.Error("Failed to create gzip reader", //nolint:contextcheck // no ctx
					log.ErrAttr(err))

				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

				return
			}

			defer func(gzipReader *gzip.Reader) { //nolint:contextcheck // no ctx
				err = gzipReader.Close()
				if err != nil {
					log.Error("Failed to close gzip reader",
						log.ErrAttr(err))
				}
			}(gzipReader)

			r.Body = io.NopCloser(gzipReader)
		}

		var respData = new(responseData)

		interceptor := &gzipResponseWriter{
			ResponseWriter: w,
			responseData:   respData,
		}

		next.ServeHTTP(interceptor, r)

		contentTypes := interceptor.Header().Values("Content-Type")

		if !shouldCompress(r, methodCompressGzip, interceptor.statusCode, interceptor.bodySize, contentTypes) { //nolint:contextcheck // no ctx
			_, err := interceptor.ResponseWriter.Write(interceptor.body)
			if err != nil {
				log.Error("Error writing response", //nolint:contextcheck // noctx
					log.ErrAttr(err))

				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}

			return
		}

		// TODO: Разобраться почему они не записываются
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Cache-Control", "no-transform")

		gzipWriter := gzip.NewWriter(w)

		defer func(gzipWriter *gzip.Writer) { //nolint:contextcheck // no ctx
			err := gzipWriter.Close()
			if err != nil {
				log.Error("Error closing gzip writer",
					log.ErrAttr(err))
			}
		}(gzipWriter)

		_, err := gzipWriter.Write(interceptor.body)
		if err != nil {
			log.Error("Error writing response", //nolint:contextcheck // noctx
				log.ErrAttr(err))

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}
	})
}

func shouldDecompress(r *http.Request, methodCompress string) bool {
	if r.Body == http.NoBody {
		return false
	}

	acceptEncodingsJoint := strings.Join(r.Header.Values("Content-Encoding"), ", ")
	if !strings.Contains(acceptEncodingsJoint, methodCompress) {
		return false
	}

	log.Debug("should to decompress")

	return true
}

func shouldCompress(r *http.Request, methodCompress string, statusCode int, bodySize int, contentTypes []string) bool {
	const maxStatusCodeForCompress = 300

	if statusCode >= maxStatusCodeForCompress {
		return false
	}

	const minSizeForCompress = 16

	if bodySize < minSizeForCompress {
		return false
	}

	contentTypesJoint := strings.Join(contentTypes, ", ")
	if strings.Contains(contentTypesJoint, "image") ||
		strings.Contains(contentTypesJoint, "video") ||
		strings.Contains(contentTypesJoint, "audio") ||
		strings.Contains(contentTypesJoint, "zip") ||
		strings.Contains(contentTypesJoint, "pdf") ||
		strings.Contains(contentTypesJoint, "text/plain") {
		return false
	}

	acceptEncodingsJoint := strings.Join(r.Header.Values("Accept-Encoding"), ", ")
	if !strings.Contains(acceptEncodingsJoint, methodCompress) {
		return false
	}

	cacheControlsJoint := strings.Join(r.Header.Values("Cache-Control"), ", ")
	if strings.Contains(cacheControlsJoint, "no-transform") {
		return false
	}

	log.Debug("should to compress")

	return true
}
