package wgwatcher

type PersistorDomain interface {
	Save(tag string, b []byte) error
	Load(tag string) ([]byte, error)
}
