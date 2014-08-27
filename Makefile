all:
	@true

.PHONY: install
install:
	mkdir -p /usr/share/templates
	install -Dm644 templates/menustyle.tmpl /usr/share/templates/menustyle.tmpl
