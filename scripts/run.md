# Script manual

## Start uranus

### 1. Deploy executables

Before you start the experiment, all the required executables should be copied from the build directory to the `scripts` folder, which is prepared for experiments. For example, in the experiment for uranus, you should copy `../build/logger`, `../build/store`, `../build/uranus` and `../build/uranus-latency` to this folder.

### 2. Generate shard config file

After you deploy all the executables in your directionary, you shuold generate shard config file like **_cross_shard_config.json_**.

- `ip_in_use`: This field contains a map from your available IP addresses to their proportion. Please write all the available IPs in the `ip_in_use` field and assign a proportion to each of them. For example, if you want to allocate two cores to each node, you can assign a proportion of 2 to the four-core server with the IP address 192.168.200.11.

- `shard`: This field contains basic information about each shard, as well as the neighbor information of the shards. Please fill in the configuration information of each shard, including: `chain_id`, the name of the shard, which should be unique; `peer_num`, the number of nodes in this shard; `related_shards`, all related shards of this shard; `is_ishard`, whether the shard is an I shard; `key_range`, the primary key range of the shard.

> **We have provided a batch of pre-written JSON files in the scripts folder, each containing different numbers of shards, numbers of shard nodes, and ω parameters. For example, `11*40*0.3.json` indicates that this is a configuration file for 11 shards, where each shard contains 40 nodes and ω = 0.3.**

### 3. Generating configuration files for each node

Generate configuration files for each node by `./uranus --method=generate --config=./11_40_0.3.json --root=./mytestnet`. You can replace the `config` field with any other JSON configuration file, and replace the `root` field with any output folder path you want (if the folder itself does not exist, a folder will be created).

Our program will sequentially assign an IP address to each node, indicating that the node should be deployed on this IP address. Each node will be assigned an increasing sequence number, such as the 40 nodes for the first shard b1 will be assigned node numbers from node1 to node40, and so on.

After generating all files, you can find some directionaries named by IP addresses in your root directionary.

### 4. Deploy the configuration files to the cluster

You should distribute your folder prepared for experiment to all the servers, and you can specify any location.

### 5. Start all the nodes

You can use `start-uranus.sh` in each server to all its nodes. For example, to start nodes deployed in 192.168.200.11 at 18:00, you can use the command `bash start.sh ./mytestnet/192.168.200.11 18:00`, where `./scripts/mytestnet` can be replaced by any of your own directionary where you deploy your configuration files in the 3rd step.

- The first input parameter `target_folder` represents the path of the directory named with the IP address that you deploy on this server

- The second input parameter `start_time` indicates the time you want the experiment to start.

All nodes will not start until this time point, which makes it convenient for you to start all nodes.

```bash
#!/bin/bash
# This is start-uranus.sh
target_folder=$1
start_time=$2

for subfolder in "$target_folder"/*; do
    if [ -d "$subfolder" ]; then
        (./uranus --root="$subfolder" --start-time="$start_time" > "$subfolder"/.out 2>&1 &)
    fi
done
```

### 6. Stop all the nodes

You should wait for minutes until the experiment finishes. You can use command `killall uranus` to stop nodes in each server.

### 7. Calculate TPS and abort rate

In the experiment, nodes will record log files through blocklogger and store them in files like `./mytestnet/192.168.200.11/node1/node1-blocklogger-brief.txt`.

You can download these files and obtain the throughput and average abort rate for each shard by using the command `./logger ./outputs/node1-blocklogger-brief.txt`, if you have downloaded such a file into the folder named outputs.

In the test load, we assign the same number of cross-shard transactions to each shard, so the abort rate of cross-shard transactions can be calculated by averaging. For the abort rate of all transactions, it is calculated by weighted summation using the proportion of cross-shard transactions.

### 8. Calculate latency

To calculate average transaction commit latency, you can use `./uranus-latency ./mytestnet/192.168.200.11/node1/ b1` in each node and calculate it by weighted summation using the proportion of cross-shard transactions.

- The first parameter is the root directionary of node. For example, `./mytestnet/192.168.200.11/node1` means the root directionary of `node1`, which is deployed in server at 192.168.200.11.

- The second parameter is the shard which the node belongs to.

### 9. Remove database and log files for another round of experiment

You maynot wish to deploy all files whenever starting a new round of experiment. You can use `bash remove.sh` to remove database and all its responding log files on each node.

- The first input parameter `target_folder` represents the path of the directory named with the IP address that you deploy on this server.

```bash
#!/bin/bash

# remove.sh

target_folder=$1

for subfolder in "$target_folder"/*; do
    if [ -d "$subfolder" ]; then
        nodeName=$(basename $subfolder)
        rm -r $subfolder/database
        rm $subfolder/$nodeName-blocklogger-brief.txt
        rm $subfolder/$nodeName-blocklogger.txt
        mkdir $subfolder/database
    fi
done
```

#### 10. Calculate the data storage overhead

The executable `store` is used to calculate the storage overhead of each node. In uranus, there is 2 types of database.

- `consensus.db` stores the blocks.
- `abci.minibank.db` stores the accounts and their balances.

For example, `./store ./mytestnet/192.168.200.11/node1/database/consensus.db` can be used to calculate the database used for consensus of node1 deployed in 192.168.200.11.

#### 11. Change the experiment setups

For different setups, you can modify `uranus/main/start.go`. You can change the step size of each transfer transaction, the transaction arrival rate or the proportion of cross-shard transactions in this file.

- `transferSize` parameter specifies the step size.
- `mustLen` parameter specifies the number of bytes of the transaction.
- `accountNumTotal` parameter specifies the total number of virtual accounts in the system.
- `commit_rates` specifies the total load of the system.
- `cross_shard_rates` specifies the proportion of cross-shard transactions in the system.
- `txTotals` specifies the size of the experimental set.

After you modify this file, you should use `make build-all` to reconstruct the executable files. Then you can turn to the 1st step and carry out the expriment again.

## Start pyramid

Pyramid and Uranus use the same configuration file; the only difference is that the startup command of pyramid is stored in `pyramid/main/start.go`, its executable program is named `pyramid`, and it uses `pyramid-latency` to calculate the latency. You can start pyramid using the same method as Uranus, and test the throughput, latency, and abort rate.

## Start rapid

We implemented the transfer mechanism of Rapid based on Pyramid. You can use JSON files like `./rapid11.json` to generate configuration files for each node of RapidChain in the 3rd step.

The executable file is named `rapid`, and `rapid-latency` is used to calculate the latency. You can modify the system of RapidChain by `rapid/main/start.go` as well.
