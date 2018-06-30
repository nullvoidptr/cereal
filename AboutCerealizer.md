# About Cerealizer #

The following is the documentation provided by `pydoc cerealizer`. It is provided
as a reference for the data format `cereal` decodes.

```
NAME
    cerealizer - Cerealizer -- A secure Pickle-like module

FILE
    /usr/lib/python2.7/dist-packages/cerealizer/__init__.py

DESCRIPTION
    The interface of the Cerealizer module is similar to Pickle, and it supports
    __getstate__, __setstate__, __getinitargs__ and __getnewargs__.
    
    Cerealizer supports int, long, float, bool, complex, string, unicode, tuple, list, set, frozenset,
    dict, old-style and new-style class instances. C-defined types are supported but saving the C-side
    data may require to write e.g. a specific Handler or a __getstate__ and __setstate__ pair.
    Objects with __slots__ are supported too.
    
    You have to register the class you want to serialize, by calling cerealizer.register(YourClass).
    Cerealizer can be considered as secure AS LONG AS the following methods of 'YourClass' are secure:
      - __new__
      - __del__
      - __getstate__
      - __setstate__
      - __init__ (ONLY if __getinitargs__ is used for the class)
    
    These methods are the only one Cerealizer may call. For a higher security, Cerealizer maintains
    its own reference to these method (exepted __del__ that can only be called indirectly).
    
    Cerealizer doesn't aim at producing Human-readable files. About performances, Cerealizer is
    really fast and, when powered by Psyco, it may even beat cPickle! Although Cerealizer is
    implemented in less than 500 lines of pure-Python code (which is another reason for Cerealizer
    to be secure, since less code means less bugs :-).
    
    Compared to Pickle (cPickle):
     - Cerealizer is secure
     - Cerealizer achieves similar performances (using Psyco)
     - Cerealizer requires you to declare the serializable classes
    
    Compared to Jelly (from TwistedMatrix):
     - Cerealizer is faster
     - Cerealizer does a better job with object cycles, C-defined types and tuples (*)
     - Cerealizer files are not Human readable
    
    (*) Jelly handles them, but tuples and objects in a cycle are first created as _Tuple or
    _Dereference objects; this works for Python classes, but not with C-defined types which
    expects a precise type (e.g. tuple and not _Tuple).
    
    
    
    IMPLEMENTATION DETAILS
    
    GENERAL FILE FORMAT STRUCTURE
    
    Cerealizer format is simple but quite surprising. It uses a "double flat list" format.
    It looks like that :
    
      <magic code (currently cereal1)>\n
      <number of objects>\n
      <classname of object #0>\n
      <optional data for creating object #0 (currently nothing except for tuples)>
      <classname of object #1>\n
      <optional data for creating object #1 (currently nothing except for tuples)>
      [...]
      <data of object #0 (format depend of the type of object #0)>
      <data of object #1 (format depend of the type of object #1)>
      [...]
      <reference to the 'root' object>
    
    As you can see, the information for a given object is splitted in two parts, the first one
    for object's class, and the second one for the object's data.
    
    To avoid problems, the order of the objects is the following:
    
      <list, dict, set>
      <object, instance>
      <tuple, sorted by depth (=max number of folded tuples)>
    
    Objects are put after basic types (list,...), since object's __setstate__ might rely on
    a list, and thus the list must be fully loaded BEFORE calling the object's __setstate__.
    
    
    DATA (<data of object #n> above)
    
    The part <data of object #n> saves the data of object #n. It may contains reference to other data
    (see below, in Cerealizer references include reference to other objects but also raw data like int).
    
     - an object           is saved by :  <reference to the object state (the value returned by
                                          object.__getstate__() or object.__dict__)>
                                          e.g. 'r7\n' (object #7 being e.g. the __dict__).
    
     - a  list or a set    is saved by :  <number of item>\n
                                          <reference to item #0>
                                          <reference to item #1>
                                          [...]
                                          e.g. '3\ni0\ni1\ni2\n' for [0, 1, 2]
    
     - a  dict             is saved by :  <number of item>\n
                                          <reference to value #0>
                                          <reference to key #0>
                                          <reference to value #1>
                                          <reference to key #1>
                                          [...]
    
    
    REFERENCES (<reference to XXX> above)
    
    In Cerealizer a reference can be either a reference to another object being serialized in the
    same file, or a raw value (e.g. an integer).
     - an int              is saved by e.g. 'i187\n'
     - a  long             is saved by e.g. 'l10000000000\n'
     - a  float            is saved by e.g. 'f1.07\n'
     - a  bool             is saved by      'b0' or 'b1'
     - a  string           is saved by e.g. 's5\nascii' (where 5 is the number of characters)
     - an unicode          is saved by e.g. 'u4\nutf8'  (where 4 is the number of characters)
     - an object reference is saved by e.g. 'r3\n'      (where 3 means reference to object #3)
     -    None             is saved by      'n'

PACKAGE CONTENTS
    datetime_handler

FUNCTIONS
    dump(obj, file, protocol=0)
        dump(obj, file, protocol = 0)
        
        Serializes object OBJ in FILE.
        FILE should be an opened file in *** binary *** mode.
        PROTOCOL is unused, it exists only for compatibility with Pickle.
    
    dumps(obj, protocol=0)
        dumps(obj, protocol = 0) -> str
        
        Serializes object OBJ and returns the serialized string.
        PROTOCOL is unused, it exists only for compatibility with Pickle.
    
    freeze_configuration()
        freeze_configuration()
        
        Ends Cerealizer configuration. When freeze_configuration() is called, it is no longer possible
        to register classes, using register().
        Calling freeze_configuration() is not mandatory, but it may enforce security, by forbidding
        unexpected calls to register().
    
    load(file)
        load(file) -> obj
        
        De-serializes an object from FILE.
        FILE should be an opened file in *** binary *** mode.
    
    loads(string)
        loads(file) -> obj
        
        De-serializes an object from STRING.
    
    register(Class, handler=None, classname='')
        register(Class, handler = None, classname = "")
        
        Registers CLASS as a serializable and secure class.
        By calling register, YOU HAVE TO ASSUME THAT THE FOLLOWING METHODS ARE SECURE:
          - CLASS.__new__
          - CLASS.__del__
          - CLASS.__getstate__
          - CLASS.__setstate__
          - CLASS.__getinitargs__
          - CLASS.__init__ (only if CLASS.__getinitargs__ exists)
        
        HANDLER is the Cerealizer Handler object that handles serialization and deserialization for Class.
        If not given, Cerealizer create an instance of ObjHandler, which is suitable for old-style and
        new_style Python class, and also C-defined types (although if it has some C-side data, you may
        have to write a custom Handler or a __getstate__ and __setstate__ pair).
        
        CLASSNAME is the classname used in Cerealizer files. It defaults to the full classname (module.class)
        but you may choose something shorter -- as long as there is no risk of name clash.

DATA
    __all__ = ['load', 'dump', 'loads', 'dumps', 'freeze_configuration', '...

```