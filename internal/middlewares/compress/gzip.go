package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"sync"
)

// Пулы для переиспользования gzip readers и writers
var (
	gzipWriterPool = sync.Pool{
		New: func() interface{} {
			// Используем уровень сжатия по умолчанию для баланса скорости и качества
			w, _ := gzip.NewWriterLevel(nil, gzip.DefaultCompression)
			return w
		},
	}
	gzipReaderPool = sync.Pool{
		New: func() interface{} {
			return new(gzip.Reader)
		},
	}
)

type compressWriter struct {
	w              http.ResponseWriter
	zw             *gzip.Writer
	wroteHeader    bool
	shouldCompress bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	zw := gzipWriterPool.Get().(*gzip.Writer)
	zw.Reset(w)

	return &compressWriter{
		w:  w,
		zw: zw,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if !c.wroteHeader {
		c.WriteHeader(http.StatusOK)
	}

	if !c.shouldCompress {
		return c.w.Write(p)
	}

	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if c.wroteHeader {
		return
	}
	c.wroteHeader = true

	contentType := c.w.Header().Get("Content-Type")
	c.shouldCompress = statusCode < 300 && isAllowCompressContentType(contentType)

	if c.shouldCompress {
		c.w.Header().Set("Content-Encoding", "gzip")
		// ВАЖНО: Удаляем Content-Length, так как размер изменится
		c.w.Header().Del("Content-Length")
	}

	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	if c.zw == nil {
		return nil
	}

	var err error
	if c.shouldCompress {
		err = c.zw.Close()
	}

	// Возвращаем writer в пул для переиспользования
	gzipWriterPool.Put(c.zw)
	c.zw = nil

	return err
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr := gzipReaderPool.Get().(*gzip.Reader)
	if err := zr.Reset(r); err != nil {
		gzipReaderPool.Put(zr)
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	var errs []error

	if err := c.r.Close(); err != nil {
		errs = append(errs, err)
	}

	if c.zr != nil {
		if err := c.zr.Close(); err != nil {
			errs = append(errs, err)
		}
		// Возвращаем reader в пул для переиспользования
		gzipReaderPool.Put(c.zr)
		c.zr = nil
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
