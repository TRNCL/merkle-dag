package merkledag

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"path/filepath"
	"strings"
)

// Hash2File 函数接收一个KVStore和一个哈希值，以及一个文件路径，返回该文件的内容。
func Hash2File(store KVStore, hash []byte, path string, hp HashPool) ([]byte, error) {
	// 从KVStore中获取数据
	data, err := store.Get(hash)
	if err != nil {
		return nil, err
	}

	// 解析路径
	dir, file := filepath.Split(path)
	if dir == "" {
		dir = "."
	}

	// 检查数据是否为目录
	var obj Object
	dec := gob.NewDecoder(bytes.NewReader(data))
	err = dec.Decode(&obj)
	if err == nil {
		// 数据为目录，解析Links
		for _, link := range obj.Links {
			if link.Name == file {
				// 如果是当前路径的文件，递归调用
				if dir == "." {
					return Hash2File(store, link.Hash, ".", hp)
				} else {
					return Hash2File(store, link.Hash, dir, hp)
				}
			}
		}
		// 如果没有找到对应的文件，返回错误
		return nil, fmt.Errorf("file not found in directory: %s", path)
	}

	// 如果数据不是目录，且路径是当前目录，则返回数据
	if dir == "." && file == "" {
		return data, nil
	}

	// 如果数据不是目录，但路径不是当前目录，返回错误
	return nil, fmt.Errorf("file not found: %s", path)
}
