package storage

type IStorage interface {
	Save(string, string)
	Load(string) (string, error)
}
