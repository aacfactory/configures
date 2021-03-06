/*
 * Copyright 2021 Wang Min Xiang
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * 	http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package configures

import (
	"bytes"
	"fmt"
	"github.com/aacfactory/json"
	"github.com/goccy/go-yaml"
	"strings"
)

type RetrieverOption struct {
	Active string
	Format string
	Store  Store
}

func NewRetriever(option RetrieverOption) (retriever *Retriever, err error) {
	format := strings.ToUpper(strings.TrimSpace(option.Format))
	if format == "" || !(format == "JSON" || format == "YAML") {
		err = fmt.Errorf("create config retriever failed, format is not support")
		return
	}
	store := option.Store
	if store == nil {
		err = fmt.Errorf("create config retriever failed, store is nil")
		return
	}
	retriever = &Retriever{
		active: strings.ToUpper(strings.TrimSpace(option.Active)),
		format: format,
		store:  store,
	}
	return
}

type Retriever struct {
	active string
	format string
	store  Store
}

func (retriever *Retriever) Get() (v Config, err error) {
	root, subs, readErr := retriever.store.Read()
	if readErr != nil {
		err = fmt.Errorf("config retriever get failed, %v", readErr)
		return
	}
	if root == nil || len(root) == 0 {
		err = fmt.Errorf("config retriever get failed, not found")
		return
	}

	if retriever.format == "JSON" {
		if !json.Validate(root) {
			err = fmt.Errorf(" config retriever get failed, invalid json content")
			return
		}
	} else if retriever.format == "YAML" {
		mapped, validErr := yaml.YAMLToJSON(root)
		if validErr != nil {
			err = fmt.Errorf("config retriever get failed, invalid yaml content")
			return
		}
		if !json.Validate(mapped) {
			err = fmt.Errorf("config retriever get failed, invalid yaml content")
			return
		}
		root = bytes.TrimSpace(mapped)
	} else {
		err = fmt.Errorf("config retriever get failed, format is unsupported")
		return
	}

	if retriever.active == "" {
		v = &config{
			raw: root,
		}
		return
	}

	if subs == nil || len(subs) == 0 {
		err = fmt.Errorf("config retriever get failed, ative(%s) is not found", retriever.active)
		return
	}

	sub, hasSub := subs[retriever.active]
	if !hasSub {
		err = fmt.Errorf("config retriever get failed, ative(%s) is not found", retriever.active)
		return
	}
	if retriever.format == "JSON" {
		if !json.Validate(sub) {
			err = fmt.Errorf(" config retriever get failed, invalid json content")
			return
		}
	} else if retriever.format == "YAML" {
		mapped, validErr := yaml.YAMLToJSON(sub)
		if validErr != nil {
			err = fmt.Errorf("config retriever get failed, invalid yaml content")
			return
		}
		if !json.Validate(mapped) {
			err = fmt.Errorf("config retriever get failed, invalid yaml content")
			return
		}
		sub = bytes.TrimSpace(mapped)
	} else {
		err = fmt.Errorf("config retriever get failed, format is unsupported")
		return
	}

	merged, mergeErr := retriever.merge(root, sub)
	if mergeErr != nil {
		err = fmt.Errorf("config retriever get failed, merge ative failed %v", mergeErr)
		return
	}

	v = merged
	return
}

func (retriever *Retriever) merge(root []byte, sub []byte) (v Config, err error) {
	if !json.Validate(root) {
		err = fmt.Errorf("merge failed, bad json content")
		return
	}
	if !json.Validate(sub) {
		err = fmt.Errorf("merge failed, bad json content")
		return
	}
	dst := json.NewObjectFromBytes(root)
	src := json.NewObjectFromBytes(sub)
	err = dst.Merge(src)
	if err != nil {
		return
	}
	v = &config{
		raw: dst.Raw(),
	}
	return
}
