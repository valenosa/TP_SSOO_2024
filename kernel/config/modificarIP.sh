#!/bin/bash

# Check and set KERNEL_HOST (if needed)
# Check and set MEM_HOST (if needed)
if [ -z "$MEM_HOST" ]; then
  echo "The MEM_HOST is not set"
  echo "Using default port localhost"
  MEM_HOST=localhost
fi

if [ -z "$CPU_HOST" ]; then
  echo "The CPU_HOST is not set"
  echo "Using default port localhost"
  CPU_HOST=localhost
fi


# Update parametro in archivo Kernel_DL.json 
sed -i "s/\"ip_memory\": .*,/\"ip_memory\": \"$MEM_HOST\",/" Kernel_DL.json
sed -i "s/\"ip_cpu\": .*,/\"ip_cpu\": \"$CPU_HOST\",/" Kernel_DL.json

# Update parametro in archivo Kernel_FS.json
sed -i "s/\"ip_memory\": .*,/\"ip_memory\": \"$MEM_HOST\",/" Kernel_FS.json
sed -i "s/\"ip_cpu\": .*,/\"ip_cpu\": \"$CPU_HOST\",/" Kernel_FS.json

# Update parametro in archivo Kernel_IO.json
sed -i "s/\"ip_memory\": .*,/\"ip_memory\": \"$MEM_HOST\",/" Kernel_IO.json
sed -i "s/\"ip_cpu\": .*,/\"ip_cpu\": \"$CPU_HOST\",/" Kernel_IO.json

# Update parametro in archivo Kernel_Mem.json 
sed -i "s/\"ip_memory\": .*,/\"ip_memory\": \"$MEM_HOST\",/" Kernel_Mem.json
sed -i "s/\"ip_cpu\": .*,/\"ip_cpu\": \"$CPU_HOST\",/" Kernel_Mem.json

# Update parametro in archivo Kernel_Plani.json
sed -i "s/\"ip_memory\": .*,/\"ip_memory\": \"$MEM_HOST\",/" Kernel_Plani.json
sed -i "s/\"ip_cpu\": .*,/\"ip_cpu\": \"$CPU_HOST\",/" Kernel_Plani.json

# Update parametro in archivo Kernel_SE.json 
sed -i "s/\"ip_memory\": .*,/\"ip_memory\": \"$MEM_HOST\",/" Kernel_SE.json
sed -i "s/\"ip_cpu\": .*,/\"ip_cpu\": \"$CPU_HOST\",/" Kernel_SE.json
