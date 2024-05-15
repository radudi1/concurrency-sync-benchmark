About
=====

This application tries to benchmark performance of different concurrency synchronization methods available in Go. It measures time per operation using mutexes, rw-mutexes, unbuffered channels, buffered channels and atomic primitives. It launches an equal number of reader and writer goroutines. The writers increment an int64 by 1 and the reader reads the value.

Usage
=====

Download, compile and run with:

`./concurrency-sync-benchmark <millions operations per run> <maximum number of workers>`

- the first argument it's the number of operations per each iteration in millions; this many increments and this many reads will be executed on each iteration
- the second argument if maximum number of gouroutines - the same number for readers and writers; the application will iterate starting from 1 reader and 1 writer and than gradually increase up to this argument but excluding iterations in which the number of operations can't be equally divided to workers
  
Results
=======
On a Ryzen 3600X (6 cores, 12 threads) on Linux 6.8 with this command line:

`./concurrency-sync-benchmark 20 65536`

I got the following results:

Workers|Mutex(ns)|RWMutex(ns)|Channels(ns)|BuffChannels|Atomic(ns)|
-------|---------|-----------|------------|------------|----------|
1|36|2507|1160|47|4| 
2|118|1109|1427|88|8| 
4|300|611|1430|190|7| 
5|224|252|831|267|7| 
8|281|694|731|355|11| 
10|152|171|696|271|7| 
16|89|557|609|280|7| 
20|83|507|770|274|7| 
25|82|308|607|267|7| 
32|82|161|609|283|7| 
40|82|113|602|276|7| 
50|82|99|592|283|7| 
64|82|101|603|283|7| 
80|83|101|604|272|7| 
100|84|101|606|284|7| 
125|84|102|606|283|7| 
128|84|102|603|276|7| 
160|85|103|607|281|7| 
200|85|104|608|283|7| 
250|85|105|608|274|7| 
256|86|105|607|283|7| 
320|86|106|609|282|7| 
400|87|107|609|272|7| 
500|87|108|609|283|7| 
625|87|109|611|281|7| 
640|87|110|610|275|7| 
800|89|110|611|283|7| 
1000|87|111|611|283|7| 
1250|87|112|610|275|7| 
1280|87|112|608|284|7| 
1600|87|113|611|269|7| 
2000|87|113|608|275|7| 
2500|87|115|606|283|7| 
3125|87|113|602|284|7| 
3200|87|114|608|275|7| 
4000|87|115|606|282|7| 
5000|87|114|606|282|7| 
6250|87|114|605|278|7| 
6400|87|115|604|282|7| 
8000|88|114|604|282|7| 
10000|88|114|602|275|7| 
12500|88|115|628|284|7| 
15625|88|114|603|280|7| 
16000|88|115|602|275|7| 
20000|89|116|603|282|7| 
25000|88|115|604|283|7| 
31250|89|116|603|278|7| 
32000|89|117|606|284|7| 
40000|89|116|602|284|7| 
50000|90|115|602|282|7| 
62500|91|117|599|285|7|

Observations
============
- The clear winner (at least in my experiment) is atomic primitives. These are blazingly fast compared to anything else and give consistent results apart from very few glitches
- The infamous mutex comes (pretty much) second especially in high concurrency but an order of magnitude slower than atomic; still much faster than alternatives
- The read-write mutex would be third but only for high concurrency; in low concurrency is actually by far the worst. You should remember that this benchmark has 1:1 readers and writers. In a scenarios where there are multiple readers and very few writers the results might be seriously different.
- Buffered channels would probably be fourth for high concurrency but even better in very low concurrency
- Unbuffered channels performs the worst - about 2-3 times worse than buffered channels in high concurrency and much worse for very low concurrency

There are interesting things happening in the first part of the table. Leaving aside atomic primitives - which don't seem to be affected by anything - we see this:

- Time per operation for mutexes first increases and than decreases and than it becomes pretty stable
- Timings for buffered channels increases until it becomes stable
- Unbuffered channels and read-write mutexes get horrible results for low concurrency and decrease as the number of goroutines go up until they become stable as well

Stabilization actually occurs at about 10 - 16 goroutines which coincidentally corresponds to (hardwarer) thread count of the CPU. My assumption is that poor results of low concurrency has something to do with the dynamic frequency scaling of the CPU, which I didn't disable because I wanted results as close to real-world as possible.

Cheatsheet
==========

Use atomic primitives if:
- the resources you're trying to protect are of basic types supported by the atomic package
- and you want best performance possible no matter what
- and you don't care much about the Go mantra, don't go to Go church and generally you don't give a ... about anything except performance

Use simple mutexes if:
- the resource you're protecting is not supported by the atomic package but small in size (don't use mutex on a map with a billion items - you'll see why)
- and number of readers and writers are pretty balanced but not necessarily equal
- or you want to protect a larger portion of code (why would you? don't! it's evil! seriously!)
- and concurrency is either very low or pretty big (not just a few goroutines kind of thing)
- and you still want the damn performance
- and you stil don't give a ... about... - I think you get the picture

Use read-write mutexes if:
- if everything said for mutexes apply except that you have a ridiculously high number of reads and very few writes - I still have to test this though

Use buffered channels if:
- the resource you're protecting is not supported by atomic
- or you to return from function as soon as possible and the resource you can use in background won't be available anymore and you don't mind to deep-copy it and the possible caveats of this
- or you favour consistency and code elegance over pure performance
- or you do go to Go church
- and you have equal amounts of reads and writes
- and your buffer is big enough

Use unbuffered channels if:
- all said for buffered channels apply but
- for some reason you need to exchange data synchronously
- or you have memory constraints
- or you don't care about performance at all
- or... I can't think right now of anything else

Conclusion
==========

Atomic primitives are the best when it comes to performance but they only support some very basic data types and therefore most of the time they won't fit your needs. Mutexes are still the next best thing when you're sure that protected resources will still be available (eg: global variables) and if (a very big if) you don't mind the caveats (eg: you forget to unlock or unlock at the wrong moment or you lock for too long etc.). So I totally get why Go introduced the channel concept and why people advocate so much for it. It's clean, elegant, modern and fits many scenarios. But still... There's a very strong case for having and using the other synchronization methods when and where they best fit.