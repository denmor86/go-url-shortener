package storage

type IStorage interface {
	Add(string, string)
	Get(string) (string, error)
}
