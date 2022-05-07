# logAgent
A log collection tool based on kafka and tail  

一个基于kafka和tail的日志收集工具

	Instructions:

logAgent is a distributed log collection tool based on kafka and tail. Computers in different locations (kafka producers) can send their own local logs to Kafka uniformly. Then a computer (kafka consumer) takes the logs collected from various places from Kafka according to the specified topic, and integrates them into a total log file to realize log collection.

logAgent是一个基于kafka和tail编写的一个分布式日志收集工具。分处各处的计算机（kafka producer）可以把自己本地的日志统一发送到kafka里。然后一台计算机（kafka consumer）从kafka中根据指定的topic取从各处收集来的日志，整合成一个总日志文件，从而实现日志收集。

	func LogFromKafka(kafkaAddress []string, kafkaTopic string, collectedLogFilename string)
	
LogFromKafka listens to all partitions of the specified topic in kafka. When a new message is written, it automatically fetches this log and writes it to the all log file (if it does not exist, it will be automatically created). When the size of the all log file exceeds 50,000, an allLog_backup folder will be automatically created in the current execution directory. Backup and save the all log file into this folder. The backup all log file is named the backup time (accurate to microseconds). Then clear the current all log file to continue to receive new logs obtained from kafka.The parameters to be passed are the ip address and port of kafka, the specified topic, and the file name of the all log file.

LogFromKafka 在kafka中监听指定的topic的所有partition，当有新的消息写入时，自动取这条日志写入总日志文件（若不存在会自动创建）。当总日志文件的size超过50000时，会自动在当前执行目录下创建一个allLog_backup文件夹。将总日志文件备份并存入这个文件夹，备份的总日志文件名为备份的时间（精确到微秒）。然后清空当前的总日志文件，来继续接收新的从kafka中获得的日志。要传入的参数为kafka的ip地址和端口、指定的topic、总日志文件的文件名。

	func LogToKafka(logFilename string, kafkaAddress []string, kafkaTopic string)

LogToKafka uses the API provided by tail to monitor the local log file, and sends it to Kafka whenever a new log is written. The parameters to be passed are the local log file name, kafka's ip address and port, and the specified topic.

LogToKafka 用tail提供的api监听本地的日志文件，每当有新的日志写入时就把它发送到kafka中去。要传入的参数为本地的日志文件名、kafka的ip地址和端口、指定的topic。
