dist := dist
distProxy := $(dist)/proxy
distGateway := $(dist)/gateway
distFrontend := $(dist)/frontend
src := .
srcCfg := $(src)/config
srcGateway := $(src)/gateway
srcFrontend := $(src)/frontend
# Utils
deploy-proxy := ./utils/deploy-proxy
deploy-gateway := ./utils/deploy-gateway
deploy-frontend := ./utils/deploy-frontend

.PHONY: all prep proxy install-proxy gateway install-gateway update-gateway frontend install-frontend update-frontend

all: prep proxy gateway frontend

prep: | $(dist) $(distProxy) $(distGateway) $(distFrontend)

$(dist) $(distProxy) $(distGateway) $(distFrontend):
	@mkdir -p "$@"


# Proxy
proxy: $(distProxy)/nginx-proxy.conf

$(distProxy)/%.conf: $(srcCfg)/%.conf
	cp "$<" "$@"


# Gateway
gateway: prep $(addprefix $(distGateway)/, \
		therm-gateway_amd64\
		therm-gateway_armv6\
		init-gateway_systemd.conf\
		init-gateway_upstart.conf\
		proxy.my.example.com.crt\
		)

$(distGateway)/therm-gateway_amd64: $(srcGateway)/*.go
	cd "$(srcGateway)/"; go mod tidy
	cd "$(srcGateway)/"; go fmt
	cd "$(srcGateway)/"; go build -o "../$@"

$(distGateway)/therm-gateway_armv6: $(srcGateway)/*.go
	cd "$(srcGateway)/"; go mod tidy
	cd "$(srcGateway)/"; go fmt
	cd "$(srcGateway)/"; GOOS=linux GOARCH=arm GOARM=6 go build -o "../$@"

$(distGateway)/%.conf: $(srcCfg)/%.conf
	cp "$<" "$@"

$(distGateway)/proxy.my.example.com.crt: $(distProxy)/proxy.my.example.com.crt
	cp "$<" "$@"


# Frontend
frontend: prep $(addprefix $(distFrontend)/, \
		index.html\
		icon.svg\
		icon_mask.svg\
		icon.png\
		)

$(distFrontend)/%.html: $(srcFrontend)/%.html
	cp "$<" "$@"

$(distFrontend)/%.svg: $(srcFrontend)/%.svg
	cp "$<" "$@"

$(distFrontend)/%.png: $(srcFrontend)/%.png
	cp "$<" "$@"


# Deployment
install-proxy: proxy
	$(deploy-proxy) INSTALL

install-gateway: gateway
	$(deploy-gateway) INSTALL

update-gateway: gateway
	$(deploy-gateway) UPDATE

install-frontend: frontend
	$(deploy-frontend) INSTALL

update-frontend: frontend
	$(deploy-frontend) UPDATE


