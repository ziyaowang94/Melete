package minibank

/*
测试流程：
Testing process:
Create two B-shard applications and three I-shard applications, and customize the shard data range.
(2) Firstly, insertTx transaction was created in I-shard to form blocks, and then the Execution function was called respectively to create about 10000 intra-shard accounts of I-shard
(3) Randomly generate TransferTx transactions for these accounts, where: The B-shard must first PreExecutionB extract CrossShardData into the block in this Application, then PreExecutionI is called by the Application associated with the I shard, and finally Execution
Create a few manually to see if the changes are correct
Generate a large number of random transactions and test the performance.How long does it take on average for 10,000 transaction

var bv := utils.NewBitArrayFromBytes()
*/
