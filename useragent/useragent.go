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

import (
	"encoding/json"
	"io/ioutil"
)

type UserAgentsDef []UserAgentDef // http://www.user-agents.org/allagents.xml

type UserAgentDef struct {
	ID             string
	Identification string `json:"String"`
	Type           string
	Comment        string
	Link1          string
	Link2          string
}

func (uas UserAgentsDef) FilterByCategory(category string) UserAgentsDef {

	res := make(UserAgentsDef, 0)

	for _, ua := range uas {
		if ua.Type == category {
			res = append(res, ua)
		}
	}
	return res
}

func LoadFromFile(filename string) (uas UserAgentsDef, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &uas)

	return
}
