make deb:
	cd ..;\
	ln -sf packages/jessie debian ;\
	dpkg-buildpackage -rfakeroot -tc; \
	rm debian

make squeeze:
	cd ..;\
	ln -sf packages/squeeze debian ;\
	dpkg-buildpackage -rfakeroot -tc; \
	rm debian
