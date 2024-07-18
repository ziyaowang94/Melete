# rocksdb install manual

1. preparing the environment.

   ```bash
   sudo apt install build-essential
   sudo apt-get install libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev libgflags-dev
   ```

2. build rocksdb lib from source.

   ```bash
   wget https://github.com/facebook/rocksdb/archive/v6.28.2.zip
   
   unzip v6.28.2.zip && cd rocksdb-6.28.2
   
   make shared_lib
   make static_lib
   ```

3. install lib to system path.

   ```bash
   sudo make install-shared
   sudo make install-static
   
   echo "/usr/local/lib" | sudo tee /etc/ld.so.conf.d/rocksdb-x86_64.conf
   sudo ldconfig -v
   ```

4. test.

   ```bash
   wget https://github.com/facebook/rocksdb/archive/v6.28.2.zip
   
   unzip v6.28.2.zip && cd rocksdb-6.28.2
   
   make shared_lib
   make static_lib
   ```

