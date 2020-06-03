# scoped
> A simple way to remove fields from your JSON API by using `scope` tags inside the golang struct.

## Example Struct
Notice the `scope` tags for each field in the struct. Think of it as permissions...

```golang
type Example struct {
    ID           int64  `json:"id"`
    AdminOnlyStr string `json:"admin_only" scope:"admin"`
    UserOnlyStr  string `json:"user_only,omitempty" scope:"user"`
    BothStr      string `json:"both,omitempty" scope:"user,admin"`
    OmitStr      string `json:"omiter,omitempty" scope:"user,admin"`
    Hidden       string `json:"-"`
    All          string `json:"all,omitempty"`
}
```

## Usage

```golang
example := Example{
	ID:           1,
	AdminOnlyStr: "im an admin",
	UserOnlyStr:  "im a user",
	BothStr:      "im on both",
	Hidden:       "cant see this",
	All:          "should always be known",
}

out := scoped.New("admin", example)
```

Easy right? Based on the Example struct above, and the value for it, the `scoped` package
will remove values inside this struct that DO NOT have the `admin` scope.

## Scoped HTTP Handler
For ease of use, `scoped` has a http.Handler so you can just put the scope right in there.

```golang
func NewServer() {
	r := mux.NewRouter()
	r.Handle("/api/orders", scoped.Handler("admin", OrdersHandler))
	http.ListenAndServe(":8080", r)
}

func OrdersHandler(w http.ResponseWriter, r *http.Request) interface{} {
	adminOrders := database.FetchOrders()
	return adminOrders
}
```
