package blueconfig

// const (
// 	dbName = "blue_config"
// 	cname  = "nodes"
// )

// var (
// 	baseSession    *mgo.Session
// 	totalTime      time.Duration
// 	start          time.Time
// 	connectionLock sync.Mutex
// )

// func init() {
// 	connectionLock = sync.Mutex{}
// }

// type Node struct {
// 	ID          string    `json:"id" bson:"_id"`
// 	ParentID    string    `json:"parent_id" bson:"parent_id"`
// 	Label       string    `json:"lbl" bson:"lbl"`
// 	IsFolder    bool      `json:"is_folder" bson:"is_folder"`
// 	IsKey       bool      `json:"is_key" bson:"is_key"`
// 	Value       string    `json:"value" bson:"value"`
// 	Path        string    `json:"path" bson:"path"`
// 	ValueType   int       `json:"value_type" bson:"value_type"`
// 	LastUpdated time.Time `json:"last_updated" bson:"last_updated"`
// }

// type dir struct {
// 	mm *MM
// }

// func (d *dir) Close() {
// 	if d == nil || d.mm == nil || d.mm.Session == nil {
// 		return
// 	}
// 	d.mm.Close()
// }

// func NewDir() *dir {
// 	return &dir{mm: M("_default_")}
// }

// type MM struct {
// 	*mgo.Session
// 	started time.Time
// 	lbl     string
// }

// func (mm *MM) Close() {
// 	totalTime += time.Since(mm.started)
// 	fmt.Println(mm.lbl, "=>", time.Since(mm.started).String())
// 	mm.Session.Close()
// }

// func (d *dir) GetChildren(path string) []*Node {

// 	var all []*Node
// 	parentID := d.GetParentIDForPath(path)
// 	if path == "/" {
// 		parentID = "0"
// 	}

// 	if len(parentID) == 0 {
// 		return all
// 	}

// 	d.mm.DB(dbName).C(cname).Find(bson.M{"parent_id": parentID, "is_folder": true}).All(&all)
// 	return all
// }

// func (d *dir) GetParentIDForPath(path string) string {

// 	var one Node
// 	err := d.mm.DB(dbName).C(cname).Find(bson.M{"path": path, "is_folder": true}).One(&one)

// 	if err != nil {
// 		if err == mgo.ErrNotFound {
// 			return ""
// 		}
// 		panic(err)
// 	}
// 	return one.ID
// }

// func (d *dir) GetKeys(path string) []*Node {

// 	var all []*Node
// 	d.mm.DB(dbName).C(cname).Find(bson.M{"path": path, "is_key": true}).All(&all)
// 	return all
// }

// func (d *dir) RemoveKey(path, key string) {
// 	m := M("RemoveKey")
// 	defer m.Close()

// 	parentID := d.GetParentIDForPath(path)
// 	if len(parentID) == 0 {
// 		return
// 	}

// 	d.mm.DB(dbName).C(cname).Remove(bson.M{
// 		"parent_id": parentID,
// 		"is_key":    true,
// 		"lbl":       key,
// 	})
// }

// func (d *dir) SetKey(path, key, value string) {
// 	parentId := d.SetPath(path)

// 	id := blueutil.NewV4()
// 	ci, err := d.mm.DB(dbName).C(cname).Upsert(bson.M{"lbl": key, "parent_id": parentId, "is_key": true}, bson.M{
// 		"$setOnInsert": bson.M{
// 			"parent_id": parentId,
// 			"_id":       id,
// 			"is_key":    true,
// 			"is_folder": false,
// 			"path":      path,
// 		},
// 		"$set": bson.M{
// 			"value":        value,
// 			"last_updated": time.Now(),
// 		},
// 	})

// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Println("CI", ci.Matched, ci.Updated, ci.UpsertedId)
// }

// func (d *dir) SetPath(path string) string {

// 	splitted := strings.Split(path, "/")
// 	fmt.Println("Splitted", splitted)
// 	var parentID string = "0"
// 	var pathTree []string
// 	for _, pathItem := range splitted {
// 		if len(pathItem) == 0 {
// 			continue
// 		}

// 		pathTree = append(pathTree, pathItem)

// 		var one Node
// 		err := d.mm.DB(dbName).C(cname).Find(bson.M{"lbl": pathItem, "parent_id": parentID}).One(&one)
// 		if err != nil {
// 			if err == mgo.ErrNotFound {
// 				newID := blueutil.NewV4()
// 				d.mm.DB(dbName).C(cname).Insert(&Node{
// 					ID:          newID,
// 					LastUpdated: time.Now(),
// 					ParentID:    parentID,
// 					Label:       pathItem,
// 					Path:        "/" + strings.Join(pathTree, "/"),
// 					IsFolder:    true,
// 					IsKey:       false,
// 				})
// 				parentID = newID
// 				continue
// 			}
// 			panic(err)
// 		}
// 		parentID = one.ID
// 	}
// 	return parentID
// }

// func M(lbl string) *MM {
// 	connectionLock.Lock()
// 	defer connectionLock.Unlock()

// 	if baseSession != nil {
// 		return &MM{baseSession.Clone(), time.Now(), lbl}
// 	}

// 	s, err := mgo.Dial("mongodb://localhost")
// 	if err != nil {
// 		panic(err)
// 	}
// 	baseSession = s
// 	return &MM{s.Clone(), time.Now(), lbl}
// }

/*
	SetPath("/status/pinger/horizon/status","OK")
	pingerStatus := GetPath("/status/pinger/horizon")
	fmt.Println(pingerStatus.status)
	pingerKeys = GetKeys("/status/pinger/horizon") = {status:"OK"}
	hosts = GetChildren("/status/pinger")  = [Horizon]
	RemovePathWithKeys("/status/pinger/horizon") = Delete Horizon/*
*/

// SetPath("/status/loadaverage/horizon","03,05,10")
// SetPath("/status/loadaverage/troy","03,05,10")
// SetPath("/status/loadaverage/bigblue","03,05,10")
