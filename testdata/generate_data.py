#!/usr/bin/env python

# Script to generate cerealizer formated data files for use in unit testing
# 
# Usage: ./generate_data.py [<test_name>...]
#

import os
import sys

try:
    import cerealizer
except ImportError:
    print "Cannot import cerealizer. Please install"
    sys.exit(1)


# List of all classes to export by default
TESTS = ["test01", "test02", "test03"]


# Test classes should be testname = "_class"

class test01_class(object):
    '''simple class with primitive values'''
    def __init__(self):
        self.field_int = 12345
        self.field_str = "one two three four five"
        self.field_float = -1.234e9

class test02_class(object):
    '''class with list and dict values'''
    
    def __init__(self):
        self.field_list = [1, 2, 3, 4, 5, 300]
        self.field_list_float = [1.0, 4.1, 5.0, -3.20]
        self.field_list_string = ["aaaa", "bbbb", "cccc", "dddd"]
        self.field_dict = {"key1": 1234, "key2": 5678, "key3": 9012}
        #self.field_dict2 = {"int": 1234, "float": 123.45, "string": "Test String"}
        self.field_tuple = (1, 2, 3, 4)
        #self.field_tuple2 = ((5, 6), 7, ((8,9), 10), 11)

class test03_class(object):
    '''class with multiple complex types including other objects'''

    def __init__(self):
        self.field_obj1 = test01_class()
        self.field_obj2 = test02_class()


def export(test):
    filename = os.path.join(os.path.dirname(sys.argv[0]), test.lower() + ".dat")
    clsname = test + "_class"

    f = open(filename, "wb")
    try:
        cls = globals()[clsname]
    except KeyError:
        print "Test '%s' not defined." % (test)
        sys.exit(1)

    try:
        cerealizer.register(cls, classname=clsname)
        obj = cls()
        cerealizer.dump(obj, f)
        f.close()
    except Exception as e:
        print "Error exporting '%s': %s" % (test, e)
        sys.exit(1)

    print "Exported %s" % filename


if __name__ == "__main__":

    if len(sys.argv) > 1:
        tests = sys.argv[1:]
    else:
        tests = TESTS

    for t in tests:
        export(t)
