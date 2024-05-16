About
=====

This application tries to benchmark performance of different concurrency synchronization methods available in Go. It measures time per operation using mutexes, rw-mutexes, unbuffered channels, buffered channels and atomic primitives. It launches an equal number of reader and writer goroutines. The writers increment an int64 by 1 and the reader reads the value.

Usage
=====

Download, compile and run with:

`./concurrency-sync-benchmark <millions operations per run> <maximum number of workers>`

- the first argument it's the number of operations per each iteration in millions; this many increments and this many reads will be executed on each iteration
- the second argument if maximum number of gouroutines - the same number for readers and writers; the application will iterate starting from 1 reader and 1 writer and than gradually increase up to this argument but excluding iterations in which the number of operations can't be equally divided to workers
- results are all in nanoseconds per operation
- ROMutex and WOMutex are actually RWMutexes but only with readers or writers which should be the best case scenario for RLock and WLock
- ROAtomic and WOAtomic are atomic primitives with only readers or writers; they actually measure the speed of Load and Add
  
Results
=======
On a Ryzen 3600X (6 cores, 12 threads) on Linux 6.8 with this command line:

`./concurrency-sync-benchmark 20 65536`

I got the following results:

Workers|Mutex(ns)|RWMutex(ns)|ROMutex(ns)|WOMutext(ns)|Channels(ns)|BuffChannels(ns)|Atomic(ns)|ROAtomic(ns)|WOAtomic(ns)|
---|---|---|---|---|---|---|---|---|---|
1|11|1153|9|16|570|24|2|0|3| 
2|65|534|15|102|642|41|4|0|7| 
4|91|312|15|246|751|95|4|0|7| 
5|204|329|14|277|471|114|3|0|6| 
8|47|52|14|54|280|138|3|0|6| 
10|45|53|14|371|316|138|3|0|6| 
16|45|291|15|306|627|220|3|0|6| 
20|42|223|15|362|308|141|3|0|6| 
25|42|148|15|297|748|142|3|0|6| 
32|42|79|14|130|304|141|3|0|6| 
40|42|52|13|84|301|140|3|0|6| 
50|42|50|13|79|304|143|3|0|7| 
64|44|51|14|81|298|132|3|0|6| 
80|42|51|13|81|300|142|3|0|6| 
100|42|51|13|83|300|142|3|0|6| 
125|43|51|13|84|300|99|3|0|6| 
128|43|51|13|84|300|141|3|0|6| 
160|44|52|13|85|300|142|3|0|6| 
200|43|52|13|86|301|142|3|0|6| 
250|43|52|13|87|301|142|3|0|6| 
256|43|52|13|87|302|141|3|0|6| 
320|44|53|13|89|301|141|3|0|6| 
400|44|54|13|89|302|142|3|0|6| 
500|44|54|13|90|303|142|3|0|6| 
625|44|54|13|91|302|140|3|0|6| 
640|44|55|13|91|301|143|3|0|6| 
800|45|55|13|92|303|141|3|0|6| 
1000|44|55|13|93|302|142|3|0|6| 
1250|44|56|13|93|302|141|3|0|6| 
1280|44|56|13|94|303|141|3|0|6| 
1600|45|56|13|95|303|141|3|0|6| 
2000|44|56|14|94|303|141|3|0|6| 
2500|44|56|13|95|302|142|3|0|6| 
3125|44|57|13|95|301|141|3|0|6| 
3200|44|57|13|96|299|142|3|0|6| 
4000|44|56|14|95|301|142|3|0|6| 
5000|44|57|13|95|301|142|3|0|6| 
6250|45|57|13|95|300|142|3|0|6| 
6400|44|57|13|95|298|138|3|0|6| 
8000|44|57|13|95|300|144|3|0|6| 
10000|44|56|14|95|300|144|3|0|6| 
12500|45|57|13|95|300|141|3|0|6| 
15625|45|57|13|95|300|138|3|0|6| 
16000|45|55|14|96|299|144|3|0|6| 
20000|45|58|14|95|299|157|3|0|6| 
25000|45|55|14|95|298|147|3|0|6| 
31250|46|58|13|95|297|145|3|0|6| 
32000|46|49|13|96|295|157|3|0|7| 
40000|46|56|14|95|297|127|4|0|7| 
50000|47|57|14|95|297|150|4|0|6| 
62500|47|56|14|96|297|153|4|0|7|

Observations
============
- The clear winner (at least in my experiment) is atomic primitives. These are blazingly fast compared to anything else and give consistent results
- Atomic writes (actually additions) are pretty expensive compared to atomic reads; atomic reads actually have consistently 0 ns per operation; it's possible that some compiler or CPU optimization kicks in and sees that the value is never actually used nor changed
- The infamous mutex comes (pretty much) second especially in high concurrency but an order of magnitude slower than atomic; still much faster than alternatives
- The read-write mutex would be third but only for high concurrency or high read-to-write ratio; in low concurrency with equal read-write is actually by far the worst. In a scenarios where there are multiple reads and few writes the RWMutex moves clearly to second place.
- Buffered channels would probably be fourth for high concurrency but even better in very low concurrency
- Unbuffered channels perform the worst - about 2-3 times worse than buffered channels in high concurrency and much worse for very low concurrency
- As a side note timings include the worker spin-up time. So it seems that spinning up about 60k gorroutines takes close to no time at all if we look at ROAtomic unless the compiler was so smart as to detect that we weren't doing anything useful and didn't spin up any goroutines at all. Even in this scenario for other atomic operations it seems that it did spin up goroutines so we would still probably be in the less than a nanosecond territory (results remain consistent regardless of whether we spin up 2 or 60k goroutines) for launching a goroutine which seems pretty impressive.

There are interesting things happening in the first part of the table. Leaving aside atomic primitives (which don't seem to be affected by anything) and a few glitches we see this:

- Time per operation for mutexes first increases and than decreases and than it becomes pretty stable
- Timings for buffered channels increases until it becomes stable
- Unbuffered channels and read-write mutexes get horrible results for low concurrency and decrease as the number of goroutines go up until they become stable as well

In some previous tests stabilization occured at about 10 - 16 goroutines which coincidentally corresponds to (hardware) thread count of the CPU. My assumption is that poor results of low concurrency has something to do with the dynamic frequency scaling of the CPU, which I didn't disable because I wanted results as close to real-world as possible.

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
- if everything said for mutexes applies except that you have a high number of reads and few writes

Use buffered channels if:
- the resource you're protecting is not supported by atomic
- or you to return from function as soon as possible and the resource you can use in background won't be available anymore and you don't mind to deep-copy it and the possible caveats of this
- or you favour consistency and code elegance over pure performance
- or you do go to Go church
- and you have equal amounts of reads and writes
- and your buffer is big enough

Use unbuffered channels if:
- all said for buffered channels applies but
- for some reason you need to exchange data synchronously
- or you have memory constraints
- or you don't care about performance at all
- or... I can't think right now of anything else

Conclusion
==========

Atomic primitives are the best when it comes to performance but they only support some very basic data types and therefore most of the time they won't fit your needs. Mutexes are still the next best thing when you're sure that protected resources will still be available (eg: global variables) and if (a very big if) you don't mind the caveats (eg: you forget to unlock or unlock at the wrong moment or you lock for too long etc.). So I totally get why Go introduced the channel concept and why people advocate so much for it. It's clean, elegant, modern and fits many scenarios. But still... There's a very strong case for having and using the other synchronization methods when and where they best fit.

You might argue why bother at all. After all we're in nanosecond territory (most of the time). Who cares about nanoseconds? Go the Go way and just write simple elegant code that scales beautifully and consistently. However you should remember that it all adds up. Depending on your particular application you might have thousands, millions or billions of synchronization points in your code. It would add up to microseconds, milliseconds and so on. Depending on your needs you might care or not. But it's always good to know.