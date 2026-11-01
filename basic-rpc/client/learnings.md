# Go RPC Tutorial --- Learnings

A short reference for the Go concepts covered while building a small
CRUD-style RPC server.

## 1. Functions vs. methods

A **function** is called directly:

``` go
func Add(a, b int) int {
    return a + b
}

result := Add(10, 20)
```

A **method** is a function associated with a receiver type:

``` go
type Calculator struct {
    result int
}

func (c *Calculator) Add(a, b int) {
    c.result = a + b
}

calc := Calculator{}
calc.Add(10, 20)
```

The receiver `(c *Calculator)` associates `Add` with the `Calculator`
type. The receiver is not an ordinary function parameter; it is the
value on which the method operates.

Methods can be declared on many **defined types**, not only structs:

``` go
type Counter int

func (c *Counter) Increment() {
    *c++
}

var count Counter = 10
count.Increment() // count is now 11
```

You cannot define methods directly on a built-in type such as `int`.
Define your own type first. The receiver's base type must be a defined
type declared in the same package, and it cannot itself be a pointer or
interface type.

## 2. What `type` means

`type` declares a named type.

``` go
type Age int

type User struct {
    Name string
    Age  Age
}

type Speaker interface {
    Speak()
}
```

`Age` is a distinct named type with `int` as its underlying type. It is
not merely a TypeScript-style alias:

``` go
var n int = 25
var age Age = Age(n) // explicit conversion
```

A struct groups fields. A custom type based on `int` is still an
integer-like value, but it can have its own methods.

## 3. Value receivers and pointer receivers

A value receiver gets a copy of the receiver:

``` go
type Counter struct {
    value int
}

func (c Counter) Set(v int) {
    c.value = v // changes only the copy
}
```

A pointer receiver gets a pointer to the original value:

``` go
func (c *Counter) Set(v int) {
    c.value = v // changes the original Counter
}
```

Use a pointer receiver when the method needs to modify the receiver,
when copying it would be undesirable, or when the type's methods should
consistently use pointer receivers.

Go automatically takes the address of an addressable value for many
method calls:

``` go
var c Counter
c.Set(5) // shorthand for (&c).Set(5), when applicable
```

## 4. Pointer field access and `*`

Given:

``` go
type API struct {
    database []Item
}

func (a *API) Add(item Item) {
    a.database = append(a.database, item)
}
```

`a` is a `*API`. Go automatically dereferences a pointer when selecting
a struct field:

``` go
a.database
```

is equivalent to:

``` go
(*a).database
```

But this is not equivalent:

``` go
*a.database
```

Because selector access binds first, this means `*(a.database)`. Since
`database` is a slice, not a pointer, it cannot be dereferenced.

Useful distinctions:

-   `a` --- pointer to an `API`
-   `*a` --- the `API` value
-   `a.database` --- the database field (automatic dereference)
-   `(*a).database` --- explicit dereference, then field access
-   `*item` --- the value pointed to by `item`

## 5. Slices and deleting an element

Go does not provide a built-in `delete` function for slices. To remove
index `i` while preserving order:

``` go
items = append(items[:i], items[i+1:]...)
```

`append` has a variadic parameter. The `...` expands the elements of the
second slice into the variadic arguments.

Conceptually:

``` go
append(items[:i], items[i+1:]...)
```

passes the first slice as the destination and the elements of the second
slice as the values to append.

If order does not matter, a swap-with-last approach can avoid shifting
the remaining elements:

``` go
items[i] = items[len(items)-1]
items = items[:len(items)-1]
```

Be careful to check that `i` is a valid index before either operation.

## 6. Go RPC with `net/rpc`

RPC (Remote Procedure Call) lets one process invoke a method exposed by
another process. The client and server do not share Go variables or
memory; data is encoded, sent over the network, and decoded on the other
side.

A `net/rpc` service method must satisfy these requirements:

1.  The receiver type is exported (or a built-in type).
2.  The method is exported (starts with an uppercase letter).
3.  It has exactly two arguments.
4.  Both argument types must be suitable exported or built-in types for
    RPC encoding. The first argument may be a value or a pointer.
5.  The second argument must be a pointer.
6.  The method returns exactly one `error`.

Typical signature:

``` go
func (a *API) AddItem(args *AddItemArgs, reply *Item) error
```

The first argument does **not** have to be a pointer. A value argument
can also be valid:

``` go
func (a *API) AddItem(args AddItemArgs, reply *Item) error
```

The argument and reply data must be encodable by the RPC codec. With the
default Gob codec, fields that need to be transmitted must be exported.

## 7. Exported fields and Gob

An identifier beginning with an uppercase letter is exported:

``` go
type Item struct {
    Title string
    Body  string
}
```

Gob encodes exported struct fields. This will not transmit the lowercase
fields:

``` go
type Item struct {
    title string // unexported
    body  string // unexported
}
```

For data sent over RPC, use exported field names such as `Title` and
`Body`.

## 8. A small CRUD-style RPC service

The server owns the database. Keep it on the service instance rather
than in a separate global variable:

``` go
type Item struct {
    Title string
    Body  string
}

type AddItemArgs struct {
    Item Item
}

type GetByNameArgs struct {
    Title string
}

type EmptyArgs struct{}

type API struct {
    database []Item
}

func (a *API) AddItem(args *AddItemArgs, reply *Item) error {
    a.database = append(a.database, args.Item)
    *reply = args.Item
    return nil
}

func (a *API) GetByName(args *GetByNameArgs, reply *Item) error {
    for _, item := range a.database {
        if item.Title == args.Title {
            *reply = item
            return nil
        }
    }

    return errors.New("item not found")
}

func (a *API) GetDB(args *EmptyArgs, reply *[]Item) error {
    *reply = a.database
    return nil
}
```

Notice:

-   `a.database` accesses the slice owned by this API instance.
-   `*reply = item` writes the response through the reply pointer.
-   Return `nil` when the operation succeeds.
-   Return an error when the requested item does not exist.
-   `EmptyArgs` is an exported argument type for a method that needs no
    meaningful input.

A delete method can use the slice deletion idiom:

``` go
func (a *API) DeleteItem(args *GetByNameArgs, reply *Item) error {
    for i, item := range a.database {
        if item.Title == args.Title {
            *reply = item
            a.database = append(a.database[:i], a.database[i+1:]...)
            return nil
        }
    }

    return errors.New("item not found")
}
```

## 9. Register and start the RPC server

``` go
api := new(API)

if err := rpc.Register(api); err != nil {
    log.Fatal(err)
}

rpc.HandleHTTP()

listener, err := net.Listen("tcp", ":4040")
if err != nil {
    log.Fatal(err)
}

log.Fatal(http.Serve(listener, nil))
```

-   `rpc.Register(api)` registers the exported RPC methods of the
    service.
-   `rpc.HandleHTTP()` installs the RPC HTTP handler on the default HTTP
    mux.
-   `net.Listen("tcp", ":4040")` listens on port 4040 on the available
    network interfaces.
-   `http.Serve(listener, nil)` serves requests using the default HTTP
    mux.

The string expression `":" + PORT` is valid when `PORT` is a string. For
an integer port, convert it first, for example with
`strconv.Itoa(port)`.

## 10. Call the RPC methods from a client

``` go
client, err := rpc.DialHTTP("tcp", "localhost:4040")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

args := AddItemArgs{
    Item: Item{Title: "First", Body: "A first item"},
}

var reply Item

if err := client.Call("API.AddItem", &args, &reply); err != nil {
    log.Fatal(err)
}

var items []Item
if err := client.Call("API.GetDB", &EmptyArgs{}, &items); err != nil {
    log.Fatal(err)
}

fmt.Println(items)
```

The RPC method name is `"ServiceType.MethodName"`, here `"API.AddItem"`.

Always check the error returned by `client.Call`. If a call fails, the
reply should not be treated as a successful result.

## 11. Important client-server distinction

The client and server are separate processes:

-   The server's `API` instance owns the database.
-   The client's local variables are not the server's variables.
-   Calling `AddItem` sends a request to the server; the server modifies
    its own slice.
-   Calling `GetDB` asks the server to send back a copy of the current
    data.

A package-level global variable in the client is not the same variable
as a package-level global variable in the server.

## 12. Concurrency safety

RPC handlers may be invoked concurrently. If multiple requests read and
write the same slice at the same time, protect access with a mutex:

``` go
type API struct {
    mu       sync.Mutex
    database []Item
}

func (a *API) AddItem(args *AddItemArgs, reply *Item) error {
    a.mu.Lock()
    defer a.mu.Unlock()

    a.database = append(a.database, args.Item)
    *reply = args.Item
    return nil
}
```

Apply the same locking discipline to every method that reads or writes
the shared database. For read-only methods, an `sync.RWMutex` can allow
concurrent readers.

## Quick recap

  Concept               Remember
  --------------------- ----------------------------------------------------
  `type`                Declares a named type
  `struct`              Groups fields into a type
  Method receiver       Associates a method with a defined type
  `(a API)`             Value receiver; receives a copy
  `(a *API)`            Pointer receiver; can modify the original
  `a.database`          Go automatically dereferences `a` for field access
  `...` in `append`     Expands a slice into variadic arguments
  Exported identifier   Starts with an uppercase letter
  RPC reply             Must be a pointer
  RPC method return     Must be `error`
  Gob                   Transmits exported struct fields
  Server state          Lives in the server process, not the client
