package filegetter

type FileGetter interface {
	Get(filepath string) ([]byte, error)
}
