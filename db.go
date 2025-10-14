package blueconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sfi2k7/microweb"
	"go.etcd.io/bbolt"
)

/*

	path examples:
		node1
		node1/node2
		node1/node2/node3
		node1/node2/node3/node4
		node1/node2/node3/node4/node5

		each node has properties:
			node1:
				name, age, place etc


	object {
		path:'path_to_node'
		propname:'prop_name'
		propvalue:'prop_value'
	}

	setNode = node
	setNodeWithProps = node.prop {value}

	getNode (node)
	getNodeWithProps(node)

	getValue(node.prop)
	setValue(node.prop, value)
	deleteProp(node.prop)


*/

type TreeOptions struct {
	StorageLocationOnDisk string
	Port                  int
	Token                 string
}

const (
	CmdCreatePath = 99
	CmdDeleteNode = 100
)

type tree struct {
	db       *bbolt.DB
	diskpath string
	port     int
	token    string
}

func NewOrOpenTree(options TreeOptions) (*tree, error) {

	os.MkdirAll(filepath.Dir(options.StorageLocationOnDisk), 0600)

	db, err := bbolt.Open(options.StorageLocationOnDisk, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &tree{db: db, port: options.Port, token: options.Token}, nil
}

func fixpath(p string) string {
	p = strings.ReplaceAll(p, "//", "/")
	p = strings.TrimPrefix(p, "/")
	p = strings.TrimSuffix(p, "/")

	if p == "/" || p == "" || p == "root" {
		p = "root"
	} else {
		if !strings.HasPrefix(p, "root/") {
			p = "root/" + p
		}
	}

	return p
}

// prop and value will be extracted based on offset
// for example:
// node1/childnode/prop where prop is prop and offset is 1
// node1/childnode/prop/value where prop is prop and value is value and offset is 2
func parsePath(p string, offset int) (node, prop, value string, err error) {
	p = fixpath(p)

	if offset == 0 {
		return p, "", "", nil
	}

	splitted := strings.Split(p, "/")

	if len(splitted) <= offset {
		return "", "", "", errors.New("invalid offset")
	}

	head := splitted[:len(splitted)-offset]
	if len(splitted)-len(head) < offset {
		return "", "", "", errors.New("invalid offset2")
	}

	node = strings.Join(head, "/")
	tail := splitted[len(splitted)-offset:]
	prop = tail[0]

	if len(tail) > 1 {
		value = tail[1]
	}

	return node, prop, value, nil
}

func (t *tree) Close() error {
	return t.db.Close()
}

func (t *tree) rbucket(path string, offset int, fn func(b *bbolt.Bucket) error) error {

	start := time.Now()
	defer func() {
		fmt.Printf("Bucket(%s) took %s\t", path, time.Since(start))
	}()

	path = fixpath(path)
	var err error
	path, _, _, err = parsePath(path, offset)
	// fmt.Println("parsed path", path)
	if err != nil {
		return err
	}

	splitted := strings.Split(path, "/")

	return t.db.View(func(tx *bbolt.Tx) error {

		var b *bbolt.Bucket

		for _, s := range splitted {
			if b == nil {
				b = tx.Bucket([]byte(s))

			} else {
				b = b.Bucket([]byte(s))
			}

			if b == nil {
				return errors.New("path does not exist")
			}
		}

		return fn(b)
	})
}

func (t *tree) rwbucket(path string, fn func(b *bbolt.Bucket) error) error {
	start := time.Now()
	defer func() {
		fmt.Printf("Bucket(%s) took %s\t", path, time.Since(start))
	}()

	path = fixpath(path)

	splitted := strings.Split(path, "/")
	return t.db.Update(func(tx *bbolt.Tx) error {

		var b *bbolt.Bucket

		for _, s := range splitted {
			var err error
			if b == nil {
				b, err = tx.CreateBucketIfNotExists([]byte(s))

			} else {
				b, err = b.CreateBucketIfNotExists([]byte(s))
			}

			if err != nil {
				return err
			}
		}

		return fn(b)
	})
}

func (t *tree) CreatePath(p string) error {
	return t.rwbucket(p, func(b *bbolt.Bucket) error {
		return nil
	})
}

func (t *tree) DeleteNode(p string, force bool) error {
	p = fixpath(p)
	splitted := strings.Split(p, "/")
	//remove the last node and join the string
	nodetodelete := splitted[len(splitted)-1]
	splitted = splitted[:len(splitted)-1]
	p = strings.Join(splitted, "/")

	fmt.Println("going to delete", nodetodelete)
	if nodetodelete == "/" || nodetodelete == "root" || nodetodelete == "/root" {
		return errors.New("can not delete root node")
	}

	return t.rwbucket(p, func(b *bbolt.Bucket) error {

		innerb := b.Bucket([]byte(nodetodelete))
		if innerb == nil {
			return errors.New("node does not exist")
		}

		err := innerb.ForEachBucket(func(k []byte) error {
			return errors.New("node has nested nodes - must force to delete")
		})

		if err != nil && !force {
			return err
		}

		return b.DeleteBucket([]byte(nodetodelete))
	})
}

func (t *tree) SetValue(p, value string) error {
	p, prop, _, _ := parsePath(p, 1)
	return t.rwbucket(p, func(b *bbolt.Bucket) error {
		return b.Put([]byte(prop), []byte(value))
	})
}

func (t *tree) SetValues(p string, values map[string]interface{}) error {
	return t.rwbucket(p, func(b *bbolt.Bucket) error {

		for k, v := range values {
			err := b.Put([]byte(k), []byte(fmt.Sprintf("%v", v)))
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (t *tree) DeleteValue(p, prop string) error {
	return t.rwbucket(p, func(b *bbolt.Bucket) error {
		return b.Delete([]byte(prop))
	})
}

func (t *tree) GetAllProps(p string) ([]string, error) {
	var props []string
	err := t.rbucket(p, 0, func(b *bbolt.Bucket) error {
		return b.ForEach(func(k, v []byte) error {
			props = append(props, string(k))
			return nil
		})
	})
	return props, err
}

func (t *tree) GetAllPropsWithValues(p string) (map[string]string, error) {
	var props = make(map[string]string)
	err := t.rbucket(p, 0, func(b *bbolt.Bucket) error {
		return b.ForEach(func(k, v []byte) error {
			props[string(k)] = string(v)
			return nil
		})
	})
	return props, err
}

func (t *tree) GetValue(p string) (string, error) {
	var prop string
	_, pro, _, _ := parsePath(p, 1)
	prop = pro

	var value string
	err := t.rbucket(p, 1, func(b *bbolt.Bucket) error {
		value = string(b.Get([]byte(prop)))
		return nil
	})
	return value, err
}

func (t *tree) HasValue(p, prop string) (bool, error) {
	var hasValue bool
	err := t.rbucket(p, 0, func(b *bbolt.Bucket) error {
		hasValue = b.Get([]byte(prop)) != nil
		return nil
	})
	return hasValue, err
}

func (t *tree) GetNodesInPath(p string) ([]string, error) {
	var nodes []string
	err := t.rbucket(p, 0, func(b *bbolt.Bucket) error {
		return b.ForEachBucket(func(k []byte) error {
			nodes = append(nodes, string(k))
			return nil
		})
	})

	return nodes, err
}

type Packet struct {
	Path  string `json:"path"`
	Prop  string `json:"prop"`
	Value string `json:"val"`
}

type response struct {
	Error  string      `json:"error,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

func (t *tree) Serve() {
	if t.port == 0 {
		panic("http port not set")
	}

	web := microweb.New()
	web.Use(func(c *microweb.Context) bool {
		if t.token != "" && c.Query("token") != t.token {
			c.Json(response{Error: "Invalid token"})
			return false
		}
		return true
	})

	web.Get("/", func(c *microweb.Context) {

		path := c.R.URL.Path

		if strings.HasSuffix(path, "/props") {
			props, err := t.GetAllProps(strings.TrimSuffix(path, "/props"))
			if err != nil {
				c.Json(response{Error: err.Error()})
				return
			}
			c.Json(response{Result: props})
			return
		}

		if strings.HasSuffix(path, "/values") {
			propsvals, err := t.GetAllPropsWithValues(strings.TrimSuffix(path, "/values"))
			if err != nil {
				c.Json(response{Error: err.Error()})
				return
			}

			c.Json(response{Result: propsvals})
			return
		}

		if strings.HasSuffix(path, "/value") {
			value, err := t.GetValue(strings.TrimSuffix(path, "/value"))
			if err != nil {
				c.Json(response{Error: err.Error()})
				return
			}
			c.Json(response{Result: value})
			return
		}

		nodes, err := t.GetNodesInPath(path)
		if err != nil {
			c.Json(response{Error: err.Error()})
			return
		}
		c.Json(response{Result: nodes})
	})

	web.Post("/", func(c *microweb.Context) {
		path := c.R.URL.Path
		if strings.HasSuffix(path, "/save") {
			body, err := c.Body()
			if err != nil {
				c.Json(response{Error: err.Error()})
				return
			}

			var m = make(map[string]any)
			err = json.Unmarshal(body, &m)
			if err != nil {
				c.Json(response{Error: err.Error()})
				return
			}

			t.SetValues(strings.TrimSuffix(path, "/save"), m)
			c.Json(response{Result: true})
		}

		if strings.HasSuffix(path, "/set") {
			p, prop, val, err := parsePath(strings.TrimSuffix(path, "/set"), 2)
			fmt.Println("path:", p, "prop:", prop, "val:", val, "err:", err)

			restored, _ := url.JoinPath(strings.TrimSuffix(p, "/set"), prop)
			fmt.Println("restored", restored)

			err = t.SetValue(restored, val)
			if err != nil {
				c.Json(response{Error: err.Error()})
				return
			}
			c.Json(response{Result: true})
			return
		}

		if strings.HasSuffix(path, "/create") {
			err := t.CreatePath(strings.TrimSuffix(path, "/create"))
			if err != nil {
				c.Json(response{Error: err.Error()})
				return
			}

			c.Json(response{Result: true})
			return
		}

		c.Json(response{Error: "unknown error"})

		/*
			examples:
				/ create path
				/ set setvalue
				/ save multiple set values
		*/
	})

	web.Listen(t.port)
}
