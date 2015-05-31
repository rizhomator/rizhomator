/*  Copyright (C) 2015 Leopoldo Lara Vazquez.

    This program is free software: you can redistribute it and/or  modify
    it under the terms of the GNU Affero General Public License, version 3,
	  as published by the Free Software Foundation.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package fetcher

import (
	"io/ioutil"
	"os"
)

type ConstantObjectStorage interface {
	Store(key string, data []byte) error
	Exists(key string) (bool, error)
	Retrieve(key string) ([]byte, error)
}

type LocalFS struct {
	Root string
}

func (s LocalFS) path(key string) string {
	return s.Root + "/" + key + ".stored"
}

func (s LocalFS) Exists(key string) (bool, error) {
	_, err := os.Stat(s.path(key))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func (s LocalFS) Store(key string, data []byte) error {
	if ex, _ := s.Exists(key); ex {
		return nil
	}

	return ioutil.WriteFile(s.path(key), data, 0666)
}

func (s LocalFS) Retrieve(key string) ([]byte, error) {
	if ex, err := s.Exists(key); !ex {
		return nil, err
	}

	return ioutil.ReadFile(s.path(key))
}
