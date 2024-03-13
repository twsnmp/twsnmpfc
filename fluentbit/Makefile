.PHONY: all clean mac linux_windows

all: mac linux mac

mac: in_twsnmp/in_twsnmp.darwin.so

in_twsnmp/in_twsnmp.darwin.so: in_twsnmp/in_twsnmp.go
	(cd in_twsnmp;go build -buildmode=c-shared -o in_twsnmp.darwin.so .)

linux_windows: in_twsnmp/in_twsnmp.dll

in_twsnmp/in_twsnmp.dll: fluentbit_builder.txt in_twsnmp/in_twsnmp.go
	docker run -it --rm -v $(CURDIR):/go/src fluentbit_builder
	

fluentbit_builder.txt: Docker/mk.sh Docker/Dockerfile
	(cd Docker;docker build -t fluentbit_builder .)
	touch fluentbit_builder.txt

clean:
	rm -rf  ../dist/*_twsnmp.* in_twsnmp/*.so in_twsnmp/*.h in_twsnmp/*.dll 
