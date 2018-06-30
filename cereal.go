// Copyright (c) 2018 James Bowen

/*
Package cereal implements access to python cerealizer archives.

Cerealizer is a "secure pickle-like" python module that serializes python objects
into a data stream that can be saved and read for later use. It was written by
Jean-Baptiste "Jiba" Lamy and is available at:

  https://pypi.org/project/Cerealizer/

This package provides a pure Go implementation to decode and unmarshal data stored
by python programs using the cerealizer library.
*/
package cereal
