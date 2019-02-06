package fs

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/crusttech/permit/pkg/permit"
)

type (
	fs struct {
		path string
	}
)

func NewPermitStorage(path string) (*fs, error) {
	return &fs{path: path}, nil
}

func (s fs) List(query string) (ll []*permit.Permit, err error) {
	var ff []os.FileInfo

	if ff, err = ioutil.ReadDir(s.path); err != nil {
		return
	}

	ll = make([]*permit.Permit, 0)
	for _, f := range ff {
		if l, err := s.read(f.Name()); err != nil {
			return nil, err
		} else {
			ll = append(ll, l)
		}
	}

	return
}

func (s fs) Get(key string) (*permit.Permit, error) {
	return s.read(s.hash(key))
}

func (s fs) Create(p permit.Permit) error {
	fp := s.hash(p.Key)

	if s.exists(fp) {
		return errors.New("permit already exists")
	}

	return s.write(fp, p)
}

func (s fs) Extend(key string, t *time.Time) error {
	return s.update(s.hash(key), func(permit *permit.Permit) error {
		permit.Expires = t
		return nil
	})
}

func (s fs) Revoke(key string) error {
	return s.update(s.hash(key), func(permit *permit.Permit) error {
		permit.Valid = false
		return nil
	})
}

func (s fs) Enable(key string) error {
	return s.update(s.hash(key), func(permit *permit.Permit) error {
		permit.Valid = true
		return nil
	})
}

func (s fs) Delete(key string) error {
	fp := s.filepath(s.hash(key))
	if !s.exists(fp) {
		return permit.PermitNotFound
	}

	return errors.Wrap(os.Remove(fp), "could not remove permit file")
}

func (s fs) hash(key string) string {
	h := md5.New()
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

func (s fs) exists(filename string) bool {
	_, err := os.Stat(s.filepath(filename))
	return err == nil
}

func (s fs) update(fp string, cb func(*permit.Permit) error) error {
	if l, err := s.read(fp); err != nil || l == nil {
		return permit.PermitNotFound
	} else if err = cb(l); err != nil {
		return errors.New("could not update permit")
	} else if err = s.write(fp, *l); err != nil {
		return err
	}

	return nil
}

func (s fs) read(filename string) (l *permit.Permit, err error) {
	var f *os.File
	l = &permit.Permit{}

	if f, err = os.Open(s.filepath(filename)); err != nil {
		if os.IsNotExist(err) {
			return nil, permit.PermitNotFound
		}
		err = errors.Wrap(err, "could not read permit file")
		return
	}

	defer f.Close()

	if err = json.NewDecoder(f).Decode(&l); err != nil {
		err = errors.Wrap(err, "could not decode permit file")
	}

	return
}

func (s fs) write(filename string, l permit.Permit) (err error) {
	var f *os.File

	if f, err = os.Create(s.filepath(filename)); err != nil {
		return errors.Wrap(err, "could not create permit file")
	}

	defer f.Close()

	if err = json.NewEncoder(f).Encode(l); err != nil {
		err = errors.Wrap(err, "could not encode permit file")
	}

	return
}

func (s fs) filepath(filename string) string {
	return s.path + string(os.PathSeparator) + filename
}
