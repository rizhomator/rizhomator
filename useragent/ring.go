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

package useragent

type Ring struct {
	UAS UserAgentsDef
	Pos int
}

func (r *Ring) Get() (ua UserAgentDef) {
	if r.Pos < 0 || r.Pos >= len(r.UAS) {
		r.Pos = 0
	}

	ua = r.UAS[r.Pos]

	r.Pos++

	return
}

func (r *Ring) LoadFromFile(filename string) (err error) {
	r.UAS, err = LoadFromFile(filename)

	return
}
