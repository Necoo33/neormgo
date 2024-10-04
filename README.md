# Neorm Go - All-in-one mysql orm

Neorm Go is a imperative orm that makes you imperatively build and execute queries.

It includes this functionalities:

* Query Builder
* Schema Builder
* Table Builder,
* Alter Queries Builder
* User Queries Builder
* Actual sql driver

It has almost everything that an engineer can want from orm and it's Query, schema And table builders tested in different situations.

## Examples

### Initialization

```go

// import that liblary:

import (
    "github.com/Necoo33/neorm-go"
)

```

### Database Connection

```go

// ...

neorm := orm.Neorm{}

database, err := database.Connect("username:password@tcp(127.0.0.1:3306)/schema_name") // schema name not necessary

if err != nil {
// do your error checking
}

// ...

```

### Building Queries

There is some examples for building and executing CRUD queries. You can do all of them with the same instance imperatively, when you invoke `.Select()`, `.Insert()`, `.Update()` and `.Delete()` methods query building will be restarted. Less allocation, more performance.

#### SELECT query

```go

// ...

database = database.Select([]string{"id", "title", "description", "published", "likes", "comments"})
database.Table("blogs")
database.Where("likes", ">", 50)
database.And("published", "=", true)
database.Finish()

// than execute that query:

posts, err := database.Execute()

if err != nil {
// error checking    
}

// it returns rows as []map[string]interface{}, that means you can reach rows similar to php's associative array':

for i, post := range posts {
fmt.Printf("%d. row: \n", i+1)
fmt.Printf("Post title: %s\n", post["title"])
fmt.Printf("Published: %v\n", post["published"])
fmt.Printf("Likes: %d\n", post["likes"])
}

// ...

```

#### INSERT Query

```go

columns := []string{"id", "description", "title"} // all your columns

values := []interface{}{1, "lorem ipsum dolor sit amet", "consectetur adipiscing elit!"} // all your values ordinarily

insert := database.Insert(columns, values)
insert.Table("blogs")
insert.Finish()
insert.Execute()

```

#### Update Query

```go

database.Update()
database.Table("blogs")
database.Set("published", false)
database.Where("id", "=", 1)
database.Finish()
database.Execute()

```

#### Delete Query

```go

database.Delete()
database.Table("blogs")
database.Where("id", "=", 10)
database.Finish()
database.Execute()

```

### Schema Creation

Creating a schema is as simple as it is:

```go

schema := database.CreateSchema("neorm_test")
schema.Finish()

err = schema.QueryDrop()

if err != nil {
// do your error check
}

```

### Table Creation

Table creation is also easy and very readable. Such as that:

```go

// initialize the table or continue from another Neorm instance:

table := database.CreateTable("blogs").IfNotExist()

table.AddColumn("id")
table.Type("int")
table.PrimaryKey()
table.NotNull()

table.AddColumn("title")
table.Type("VARCHAR(30)")
table.NotNull()
table.Unique()

table.AddColumn("published")
table.Type("BOOLEAN")
table.Default(true)

table.AddColumn("description")
table.Type("TEXT(1000)")
table.Null()

table.AddColumn("user_id")
table.Type("TINYINT")
table.NotNull()

// make user_id column a foreign key for "id" column of "users" table:

type References struct {
Users string
}

table.ForeignKey("user_id", References{Users: "id"})

table.Finish()

// This codes create this query:

/* 

CREATE TABLE IF NOT EXIST (
    id INT PRIMARY KEY NOT NULL,
    title VARCHAR(30) NOT NULL UNIQUE,
    published BOOLEAN DEFAULT true,
    description TEXT(1000) NULL,
    user_id TINYINT NOT NULL
    FOREIGN KEY user_id REFERENCES users(id)
);

*/

```

That orm is built especially for my personal use but anyone who wants to empower themselves with neorm free to use it. Contributions or feature requests are welcome.
