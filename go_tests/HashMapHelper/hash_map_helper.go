package hash_map_helper

import "runtime"

var CONCURRENCY_FACTOR = 4
var CONCURRENCY = runtime.NumCPU() * CONCURRENCY_FACTOR