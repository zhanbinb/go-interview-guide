Go 使用 goroutine 实现并发。

channel 用于 goroutine 之间的数据通信。

无缓冲 channel 的发送和接收需要双方配合。

有缓冲 channel 可以在 buffer 未满时先完成发送，从而降低发送方和接收方之间的同步阻塞。