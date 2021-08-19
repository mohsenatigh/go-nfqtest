# Simple FW 

It is a simple FW program developed with GOLANG. This program captures packets using the Netfilter NFQUEUE mechanism and applies some simple quota policies against the traffic. 
It is possible to configure the system in both local and gateway modes. In the gateway mode, it applies the policies for each source IP individually. 

## Dependency 

The following libraries are required to build and test this system successfully. 

libnetfilter_queue : it can be installed using "apt-get install libnetfilter-queue-dev" in Ubunto 
github.com/Telefonica/nfqueue:  a light GO wrapper around the libnetfilter_queue 
github.com/google/gopacket:  for packet parsing and processing

## Build 

To build this system run the following commands

    Make clean
    Make  or Make debug

## test 

to run the unit tests run :

    go test 

## run 

to run the system, use the following command 
    
    simplefw.bin -f setting.json

## Configuration 

The configuration is a JSON formatted file. following is the list of  valid configurations
    
- max_conversation : maximum tracked conversations
- max_inactive_conversation_life_time :  remove inactive conversation after this interval 
- nfq_number :  Netfilter queue number
- gw_mode :  if true system runs in gateway mode otherwise, the system will run in local mode
- run_iptables_command : automatically add and remove related Iptables command.
- rules :list of rules in the following format 
- - name : name of rule 
- - destination : destination network  could be 0.0.0.0/0 for all or a host name
- - protocol : could be tcp,udp or any
- - usage_time :  allowable time usage 
- - usage_size :   allowable data usage

## API 

You can use the following APIs to query the different parts of the system:
- http://127.0.0.1:8080/conversations : list all the active conversations
- http://127.0.0.1:8080/provider : get the provider status

## Limitations

- This system just supports IPv4.
- Regarding the domain names, it just tracks one of the IP addresses, not all the CDNS

