Mysqldb
=======
Mysqldb is a simple `ORM` for `Go`.

### Installing Mysqldb
    go get github.com/go-sql-driver/mysql
    go get github.com/kirinlabs/mysqldb

### How to use Mysqldb?
Oepn a database connection that supports ConnectionPool

```go
var db *mysqldb.Adapter
var err error

func init(){
    db, err = mysqldb.New(
        &mysqldb.Options{
    	    User:         "root",
    	    Password:     "root",
            Host:         "127.0.0.1",
    	    Port:         3306,
    	    Database:     "test",
    	    Charset:      "utf8",
    	    MaxIdleConns: 5,
    	    MaxOpenConns: 10,
    	    Debug:        true,
        })
    if err != nil {
    	log.Panic("Connect mysql server error: ", err)
    }
}
```

Open debug log and set the log level

```go
  db.SetLogLevel(mysqldb.LOG_DEBUG)
```


### Crud Operation

```go
type Article struct {
    Id                      int	`json:"id"`
    Title                   string `json:"title"`
    Description             string `json:""`
    CreateDate      string    `json:"create_date"`
}
```

Fetch a single object

```go
var article1 Article
err := db.Table("article").First(&article1)

var article2 Article
err := db.Table("article").Where("id",">",1).First(&article2)

var article3 Article
err := db.Table("article").Where("id",">",1).Where("create_date","2019-01-01 12:00:01").First(&article2)

```

Fetch objects list

```go
var articles1 []*Article
err := db.Table("article").Where("id",">=","1").Limit(10).Find(&articles1)

var articles2 []*Article
err := db.Table("article").Where("id",">=","1").Limit(10,20).Find(&articles2)
```


Fetch result as Map

```go
list,err := db.Table("article").Where("id",">=",1).Where("age",19).Fetch() //return map[string]interface{}

list,err := db.Table("article").Where("id",">=",1).Where("age",19).FetchAll() //return []map[string]interface{}

list,err := db.Table("article").Where("id", ">=", 1).Limit(10).Fields("id", "title").FetchAll()
```

Insert with Map

```go
data := map[string]interface{}{
		"title":       "test Mysqldb one",
		"cid":         1,
		"Description": "test insert api",
		"create_date": time.Now().Format("2006-01-02 15:04:05"),
	}
num, err := db.Table("article").Insert(data)
```

Insert with Struct

```go
article := Article{}
article.Title = "test insert two"
article.CreateDate = time.Now().Format("2006-01-02 15:04:05")

num, err := db.Table("article").Insert(&article)
```

MultiInsert with Map

```go
_, err := m.Table("article").MultiInsert([]map[string]interface{}{
	{"title":"a","desp":"test one","create_date": time.Now().Format("2006-01-02 15:04:05")},
	{"title":"b","desp":"test two","create_date": time.Now().Format("2006-01-02 15:04:05")},
})
```

MultiInsert with Struct

```go
_, err := m.Table("article").MultiInsert([]*Article{
	{nil,"a","test one",time.Now().Format("2006-01-02 15:04:05")},
	{nil,"b","test two",time.Now().Format("2006-01-02 15:04:05")},
})
```

Update with Map

```go
data := map[string]interface{}{
		"id":          100,  //Primary keys can be filtered with SetPk(" id ")
		"title":       "test Mysqldb one",
		"cid":         1,
		"Description": "update",
		"create_date": time.Now().Format("2006-01-02 15:04:05"),
}
num, err := db.Table("article").Where("id", 1).SetPk("id").Update(data)  //SetPk("id"), Prevent primary key id from being updated
```

Delete

```go
num, err := db.Table("article").Where("id", 1).Delete()
num, err := db.Table("article").Delete() //will faild,where condition cannot be empty.
```

### Join Operation
The default alias for the Table is `A`, default alias of the Join table is `B`


Left join operation
```go
data, err := db.Table("article").LeftJoin("category", "A.cid=B.id").FetchAll()
data, err := db.Table("article").LeftJoin("category", "A.cid=B.id").WhereRaw("B.id is not null").FetchAll()
```

Right join operation
```go
data, err := db.Table("article").RightJoin("category", "A.cid=B.id").FetchAll()
```


### Advanced operations

Id
```go
list, err := db.Table("article").Id(1).FetchAll()  //The Id() default operation id field
list, err := db.Table("article").Id([]int{1,2,3}).FetchAll()

list, err := db.Table("article").SetPk("cid").Id([]int{1,2,3}).FetchAll() //Modify the field for the Id() operation
```

WhereIn
```go
list, err := db.Table("article").WhereIn("id", []int{1, 2, 3}).FetchAll()
```

WhereNotIn
```go
list, err := db.Table("article").WhereNotIn("id", []int{1, 2, 3}).FetchAll()
```

WhereRaw
```go
list, err := db.Table("article").Where("id", 2).WhereRaw("cid>=1 and description=''").FetchAll()
```

OrWhere
```go
list, err := db.Table("article").Where("id", 2).OrWhere("cid>=1 and description=''").FetchAll()
```

Limit
```go
list, err := db.Table("article").Limit(10).FetchAll()
list, err := db.Table("article").Limit(10,20).FetchAll()
```

GroupBy
```go
list, err := db.Table("article").Where("id", ">", 1).GroupBy("cid").FetchAll()
list, err := db.Table("article").Where("id", ">", 1).GroupBy("cid,title").FetchAll()
```

OrderBy
```go
list, err := db.Table("article").Where("id", ">", 1).OrderBy("id").FetchAll()
list, err := db.Table("article").Where("id", ">", 1).OrderBy("id desc").FetchAll()
```

Distinct
```go
list, err := db.Table("article").Distinct("cid").FetchAll()
```

Count
```go
list, err := db.Table("article").Count()
list, err := db.Table("article").Where("id",">=",100).Count()
list, err := db.Table("article").Distinct("cid").Count()
```

### Execute native SQL

Query
```go
list, err := db.Query("select * from article")
```

Exec
```go
num, err := db.Exec("update article set description='execute' where id=2")
```

### Transaction
Need to create a new Model object for transaction operations

```go
m := db.NewModel()

defer m.Close()

m.Begin()
_, err := m.Exec("update article set description='query' where id=3")
if err != nil {
	model.Rollback()
	return
}

_, err = m.Table("category").Where("id", 9).Delete()
if err != nil {
	m.Rollback()
	return
}

m.Commit()
```

### Helper method

Scan()
```go
article := &Article{}

data, _ := db.Table("article").Where("id",1).Where("title","!=","").Fetch()

db.Scan(data,&article)
```

```go
articles := make([]*Article,0)

list, _ := db.Table("article").Where("id",">",1).Where("title","!=","").FetchAll()

db.Scan(list,&articles)

for _,v:=range articles{
    log.Println(v.Title)
}
```
