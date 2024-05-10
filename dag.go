package merkledag

import (
	"hash"
	"bytes"
	"encoding/gob"
	"io"
)

type Link struct {
	Name string
	Hash []byte
	Size int
}

type Object struct {
	Links []Link
	Data  []byte
}

func Add(store KVStore, node Node, hp HashPool) ([]byte, error) {
	// 创建哈希对象
	h := hp.Get()

	// 处理文件节点
	if node.Type() == FILE {
		file := node.(File) // 类型断言为File
		data := file.Bytes()

		// 将文件数据写入哈希对象
		h.Write(data)

		// 将文件数据存储到KVStore
		err := store.Put(h.Sum(nil), data)
		if err != nil {
			return nil, err
		}

		// 返回文件内容的哈希作为Merkle Root
		return h.Sum(nil), nil
	}

	// 处理目录节点
	if node.Type() == DIR {
		dir := node.(Dir) // 类型断言为Dir
		iterator := dir.It()

		var links []Link

		// 遍历目录中的所有子节点
		for iterator.Next() {
			child := iterator.Node()
			childHash, err := Add(store, child, hp)
			if err != nil {
				return nil, err
			}

			// 构建子节点的Link信息
			link := Link{
				Name: child.Name(),
				Hash: childHash,
				Size: int(child.Size()),
			}
			links = append(links, link)
		}

		// 创建Object并序列化
		object := Object{
			Links: links,
		}
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(object)
		if err != nil {
			return nil, err
		}

		// 将Object数据写入哈希对象
		io.Copy(h, &buf)

		// 将Object数据存储到KVStore
		err = store.Put(h.Sum(nil), buf.Bytes())
		if err != nil {
			return nil, err
		}

		// 返回Object数据的哈希作为Merkle Root
		return h.Sum(nil), nil
	}

	return nil, fmt.Errorf("unknown node type: %d", node.Type())
}
