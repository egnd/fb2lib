package indexing

// func IterateZippedBooks(filePath string, logger zerolog.Logger) error {
// 	container, err := os.Open(filePath)
// 	if err != nil {
// 		return err
// 	}
// 	defer container.Close()

// 	containerReader, err := zip.OpenReader(filePath)
// 	if err != nil {
// 		return err
// 	}
// 	defer containerReader.Close()

// 	for _, book := range containerReader.File {
// 		logger := logger.With().Str("libsubitem", book.Name).Logger()

// 		if book.Method != zip.Deflate {
// 			logger.Warn().Uint16("compression", book.Method).Msg("check item compression type")
// 			continue
// 		}

// 		offset, err := book.DataOffset()
// 		if err != nil {
// 			logger.Error().Err(err).Msg("readig zip item offset")
// 			continue
// 		}

// 		flateReader := flate.NewReader(io.NewSectionReader(container, offset, int64(book.CompressedSize64)))

// 	}

// 	return nil
// }
