package fonts

import (
	"fmt"
	"io"

	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
)

// NotoSansCJKJP is the full, unmodified Japanese CJK family, including Latin and digits.
const NotoSansCJKJP = "notosanscjkjp"

// ReaderAt lets sfnt read from embedded font storage without copying both large files to the heap.
func notoFonts() ([2]*sfnt.Font, error) {
	var out [2]*sfnt.Font
	for i, name := range []string{"Regular.otf", "Bold.otf"} {
		file, err := files.Open("files/notosanscjkjp/" + name)
		if err != nil {
			return out, err
		}
		reader, ok := file.(io.ReaderAt)
		if !ok {
			file.Close()
			return out, fmt.Errorf("Noto font %s has no ReaderAt", name)
		}
		f, err := opentype.ParseReaderAt(reader)
		if err != nil {
			file.Close()
			return out, fmt.Errorf("Noto font %s: %w", name, err)
		}
		// embed.FS files have no external descriptor to close; the parsed font retains the reader.
		out[i] = f
	}
	return out, nil
}
