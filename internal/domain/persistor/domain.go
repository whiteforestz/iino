package persistor

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	tagRoot    = "iino"
	tmpPattern = "%s-*"

	defaultModeFile = 0664
	defaultModeDir  = 0775
)

type Domain struct {
	cfg Config

	tags map[string]string
}

func New(
	cfg Config,
) *Domain {
	return &Domain{
		cfg: cfg,

		tags: make(map[string]string),
	}
}

func (d *Domain) Prepare() error {
	err := os.Mkdir(d.cfg.RootPath, defaultModeDir)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("can't mk root dir: %w", err)
	}

	path, err := ioutil.TempDir("", fmt.Sprintf(tmpPattern, tagRoot))
	if err != nil {
		return fmt.Errorf("can't mk tmp root dir: %w", err)
	}

	d.tags[tagRoot] = path

	return nil
}

func (d *Domain) Clean() error {
	rootPath, found := d.tags[tagRoot]
	if !found {
		return errors.New("root tag not found")
	}

	if err := os.RemoveAll(rootPath); err != nil {
		return fmt.Errorf("can't clean tmp root dir: %w", err)
	}

	return nil
}

func (d *Domain) Save(tag string, b []byte) error {
	tagPath, err := d.prepareTag(tag)
	if err != nil {
		return fmt.Errorf("can't prepare tag: %w", err)
	}

	tmpSlug, err := generateRandomSlug()
	if err != nil {
		return fmt.Errorf("can't generate tmp slug: %w", err)
	}

	var (
		pathTmp  = filepath.Join(tagPath, tmpSlug)
		pathData = filepath.Join(d.cfg.RootPath, tag)
	)

	err = ioutil.WriteFile(pathTmp, b, defaultModeFile)
	if err != nil {
		return fmt.Errorf("can't write data file: %w", err)
	}

	err = os.Rename(pathTmp, pathData)
	if err != nil {
		return fmt.Errorf("can't move data file: %w", err)
	}

	return nil
}

func (d *Domain) Load(tag string) ([]byte, error) {
	pathData := filepath.Join(d.cfg.RootPath, tag)
	b, err := ioutil.ReadFile(pathData)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotExists
		}

		return nil, err
	}

	return b, nil
}

func (d *Domain) prepareTag(tag string) (string, error) {
	if tag == tagRoot {
		return "", errors.New("invalid tag")
	}

	if tagPath, found := d.tags[tag]; found {
		return tagPath, nil
	}

	rootTagPath, found := d.tags[tagRoot]
	if !found {
		return "", errors.New("root tag not found")
	}

	tagPath, err := os.MkdirTemp(rootTagPath, fmt.Sprintf(tmpPattern, tag))
	if err != nil {
		return "", fmt.Errorf("can't mkdir: %w", err)
	}

	d.tags[tag] = tagPath

	return tagPath, nil
}

func generateRandomSlug() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("can't read rand: %w", err)
	}

	return fmt.Sprintf("%x", b), nil
}
