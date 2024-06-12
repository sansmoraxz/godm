package godm

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"fmt"
	"io"
	"os"

	"github.com/klauspost/compress/zstd"

	"github.com/dsnet/compress/brotli"
)

var supportedEncodings = []string{"gzip", "br", "zstd", "compress", "deflate"}

func getReader(encoding string, mainFile *os.File) (io.ReadCloser, error) {
	switch encoding {
	case "gzip":
		return gzip.NewReader(mainFile)
	case "br":
		return brotli.NewReader(mainFile, nil)
	case "zstd":
		if d, err := zstd.NewReader(mainFile); err != nil {
			return nil, err
		} else {
			return d.IOReadCloser(), nil
		}
	case "compress":
		return lzw.NewReader(mainFile, lzw.LSB, 8), nil
	case "deflate":
		return flate.NewReader(mainFile), nil
	default:
		return nil, fmt.Errorf("unknown encoding: %s", encoding)
	}
}

func decompressFile(mainFile *os.File, filePath string, encoding string) error {
	mainFile.Seek(0, 0)

	gr, err := getReader(encoding, mainFile)
	if err != nil {
		return err
	}
	defer gr.Close()
	log.Info("Decompressing file...")
	fout, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fout.Close()
	if _, err = io.Copy(fout, gr); err != nil {
		return err
	}
	log.Info("Decompressed file...")
	return nil
}
