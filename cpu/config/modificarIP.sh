#!/bin/bash

# Check and set KERNEL_HOST (if needed)
if [ -z "$KERNEL_HOST" ]; then
  echo "The KERNEL_HOST is not set"
  echo "Using default port localhost"
  KERNEL_HOST=localhost
fi

# Check and set MEM_HOST (if needed)
if [ -z "$MEM_HOST" ]; then
  echo "The MEM_HOST is not set"
  echo "Using default port localhost"
  MEM_HOST=localhost
fi


# Update parametro in archivo CPU_DL-IO.json 
sed -i "s/\"ip_kernel\": .*,/\"ip_kernel\": \"$KERNEL_HOST\",/" CPU_DL-IO-FS.json
sed -i "s/\"ip_memory\": .*,/\"ip_memory\": \"$MEM_HOST\",/" CPU_DL-IO-FS.json

# Update parametro in archivo CPU_Mem.json 
sed -i "s/\"ip_kernel\": .*,/\"ip_kernel\": \"$KERNEL_HOST\",/" CPU_Mem.json
sed -i "s/\"ip_memory\": .*,/\"ip_memory\": \"$MEM_HOST\",/" CPU_Mem.json

# Update parametro in archivo CPU_Plani.json
sed -i "s/\"ip_kernel\": .*,/\"ip_kernel\": \"$KERNEL_HOST\",/" CPU_Plani.json
sed -i "s/\"ip_memory\": .*,/\"ip_memory\": \"$MEM_HOST\",/" CPU_Plani.json

# Update parametro in archivo CPU_SE.json
sed -i "s/\"ip_kernel\": .*,/\"ip_kernel\": \"$KERNEL_HOST\",/" CPU_SE.json
sed -i "s/\"ip_memory\": .*,/\"ip_memory\": \"$MEM_HOST\",/" CPU_SE.json



