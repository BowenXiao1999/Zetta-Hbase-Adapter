我来阐述一下我对内存分配器的思考（面试后简单整理了一下思路）：

输入是用户指定的内存大小，输出是符合条件的内存的地址。同时也希望能照顾好内存的碎片问题，提高效率。



首先明确我们做的是操作系统和用户中间的一层。这样可以有效地屏蔽底层的复杂度和提高性能。同时系统调用sbrk也不能太多，否则效率比较低，因此先sbrk一块比较大的连续内存。这里有很多参数选择，比如内存的固定大小。



我会把它设计为多种固定大小的内存块，每一种大小的内存块分到一个level管理（比如8kb, 16kb, 32kb等等），也就是在一个内存分配器下又划分出多种大小的内存分配器。构造十多个固定内存分配器，分配内存时根据内存大小查表，决定到底由哪个分配器负责，分配后要在头部的 header 处写上哪个分配器分配的，确保正确归还。这种设计一是减少内存分配难度（减少内存碎片），二是为实现slab算法做准备。



对于其中每一个内存分配器，我计划使用Bucket再对真正的Memory Node进行一层封装。一个内存管理器有M个Bucket，每个Bucket设计为N个Memory Node。Memory Node 是抽象意义上的内存，它包含指向一个Page（4KB）。可以把Bucket理解为4N大小的内存单元。这M个Bucket可以拆成2条链表，一条记录已分配的Bucket，一条记录未分配的Bucket。分配内存的时候去记录未分配的链表里找，并且删除加入已分配内存的链表。归还内存的时候可类推。设计Bucket的意义主要是加速分配大块内存的效率。



对于一种大小的内存分配器，使用一个数组来包含所有Bucket，Bucket在数组内的index是不会变的，不过Bucket的指向会变。每个内存分配器还需要维护已分配的Bucket头指针 + 未分配的Bucket头指针。这个index是等于Bucket在位图中的位置。所以，我们通过位图快速计算出空闲内存块的index，再通过数组就能快速定位到对应Bucket。



内存分配器做的最重要的事就是快速找到可用的内存地址。因此我们可以设计额外的数据结构（类似数据库里的索引来假设）来加速查询。比如说bitmap，可以利用它来快速定位符合长度的空闲的内存块。



之后可以实现Slab算法，即每种大小的分配器可用空间不够了不用去操作系统申请，直接从更大的分配器处申请，（比如想申请8kB，已经没有了，则去16KB的分配器去找，并且拆成2个8kB，加入8kb管理器）。这里会引入更多复杂度，但是对于申请效率有提升。



发现如果做到内存池这个级别后，碎片管理就很难做了，一时想不起更好的办法。这个我之后会不断思考。



以上就是我对管理器的一个实现思考。刚接触C++，对内存管理没有很熟悉。但是在思考这个问题的过程中，也发现了效率、空间利用率很难兼得。还是需要读更多源码，领悟更多系统的设计细节。


