#################################################

.PHONY: clean writer reader

default:	clean reader writer

clean:
	rm -f slam-writer slam-reader

writer:
	go build writer/slam-writer.go

reader:
	go build reader/slam-reader.go


