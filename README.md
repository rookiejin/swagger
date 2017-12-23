# Usage


- the usage please see [https://github.com/swaggo/swag](https://github.com/swaggo/swag)
- the differences of swaggo/swag are below !

## Install

```
    go get -u github.com/rookiejin/swagger
```

* command
```
swagger -main main.go
````

* Model definitions

```
    package somepkg

    // @def Model  <- defined Model use @def
    type Model struct{
        Id int <- support int* float* , bson.ObjectId , if it`s array please use definitions
        SomeArray []Array `swag:"Array"` <- the Array should be defined
    }
    // @def Array
    type Array struct {
        SomeStruct string
    }
```

* Param definitions
```
    // @Param fieldName InWhere Type isRequired Description
    // @Param name query string true "name of the pets"
    // @Param file formData file true  "the file to upload "
    // @Param pets body @Model true "the defined model"
    // @Param page path string false "page in path"
```

* Response definitions
```
    // @Success 200 {object} @ArticleTag "ok"
    // @Failure 400 {object} @ArticleTag "error message"
```
