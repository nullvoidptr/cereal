<img src="https://github.com/jamesbo13/cereal/raw/master/logo-512.png" width=200>

# Cereal #

Package cereal implements access to python cerealizer archives.

Cerealizer is a "secure pickle-like" python module that serializes python objects
into a data stream that can be saved and read for later use. It was written by
Jean-Baptiste "Jiba" Lamy and is available at:

  https://pypi.org/project/Cerealizer/

This package provides a pure Go implementation to decode and unmarshal data stored
by python programs using the cerealizer library.

Details on the cerealizer file format can be found [here](AboutCerealizer.md).

## Example ##

```golang
package main

import (
    "fmt"
    "ioutil"
    "log"

    "github.com/jamesbo13/cereal"
)

type SampleData struct {
    IntSlice []int
    String string
    Float float32
}

func main() {

    buf, _ := ioutil.ReadFile("filename")

    var d SampleData

    err := cereal.Unmarshal(buf, &d)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Decoded data:\n%#v\n")
}

```
