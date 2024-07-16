echo "install all libraries"
sudo apt install build-essential

sudo apt-get install libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev libgflags-dev

echo "Download the installer"
mkdir ~/download
cd ~/download
wget https://github.com/facebook/rocksdb/archive/v6.28.2.zip

unzip v6.28.2.zip

cd rocksdb-6.28.2

echo "Compile the dynamic link library"

make shared_lib && sudo make install-shared

echo "Compile the statically linked library"
make static_lib && sudo make install-static

echo "install"
sudo make install

echo "Setting environment variables"
echo "/usr/local/lib" |sudo tee /etc/ld.so.conf.d/rocksdb-x86_64.conf
make shared_lib && sudo make install-shared
sudo ldconfig -v