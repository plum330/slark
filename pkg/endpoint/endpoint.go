package endpoint

import (
	"errors"
	"net"
	"net/url"
	"strconv"
)

func Scheme(scheme string, insecure bool) string {
	if insecure {
		return scheme
	}
	return scheme + "s"
}

func ParseValidAddr(addr []string, scheme string) (string, error) {
	for _, v := range addr {
		u, err := url.Parse(v)
		if err != nil {
			return "", err
		}
		if u.Scheme == scheme {
			return u.Host, nil
		}
	}
	return "", errors.New("scheme not found")
}

func ParseAddr(ln net.Listener, address string) (string, error) {
	_, port, err := net.SplitHostPort(address)
	if err != nil && ln == nil {
		return "", err
	}
	if ln != nil {
		tcpAddr, ok := ln.Addr().(*net.TCPAddr)
		if !ok {
			return "", errors.New("parse addr error")
		}
		port = strconv.Itoa(tcpAddr.Port)
	}

	is, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	index := int(^uint(0) >> 1)
	ips := make([]net.IP, 0)
	for _, i := range is {
		if (i.Flags & net.FlagUp) == 0 {
			continue
		}
		if i.Index >= index && len(ips) != 0 {
			continue
		}

		addr, e := i.Addrs()
		if e != nil {
			continue
		}
		for _, a := range addr {
			var ip net.IP
			switch at := a.(type) {
			case *net.IPAddr:
				ip = at.IP
			case *net.IPNet:
				ip = at.IP
			default:
				continue
			}

			ipBytes := net.ParseIP(ip.String())
			if !ipBytes.IsGlobalUnicast() || ipBytes.IsInterfaceLocalMulticast() {
				continue
			}
			index = i.Index
			ips = append(ips, ip)
			if ip.To4() != nil {
				break
			}
		}
	}
	var host string
	if len(ips) != 0 {
		host = net.JoinHostPort(ips[len(ips)-1].String(), port)
	}
	return host, nil
}

func ParseScheme(endpoints []string) (map[string]string, error) {
	mp := make(map[string]string)
	for _, endpoint := range endpoints {
		u, err := url.Parse(endpoint)
		if err != nil {
			return nil, err
		}
		mp[u.Port()] = u.Scheme
	}
	return mp, nil
}
