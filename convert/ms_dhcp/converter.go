package ms_dhcp

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	dhcpv1 "beryju.org/kube-dhcp/api/v1"
	"github.com/go-logr/logr"
	"github.com/gosimple/slug"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Converter struct {
	in  DHCPServer
	out string
	l   logr.Logger
}

func New(input, output string) (*Converter, error) {
	x, err := ioutil.ReadFile(input)
	if err != nil {
		return nil, err
	}
	var dhcps DHCPServer
	err = xml.Unmarshal(x, &dhcps)
	if err != nil {
		return nil, err
	}
	s, err := os.Stat(output)
	if err != nil {
		err = os.MkdirAll(output, os.ModeSticky|os.ModePerm)
		if err != nil {
			return nil, err
		}
	} else {
		if !s.IsDir() {
			return nil, fmt.Errorf("output path is not a directory")
		}
	}
	return &Converter{
		in:  dhcps,
		out: output,
	}, nil
}

func (c *Converter) Run() {
	for _, scope := range c.in.IPv4.Scopes.Scope {
		c.convertScope(scope)
	}
}

func (c *Converter) convertScope(sc Scope) {
	// Build CIDR
	m := net.IPMask(net.ParseIP(sc.SubnetMask).To4())
	ones, _ := m.Size()
	_, cidr, err := net.ParseCIDR(fmt.Sprintf("%s/%d", sc.ScopeId, ones))
	if err != nil {
		log.Println("failed to parse cidr")
		return
	}
	// Build lease duration
	// saved as days:hours:minutes
	// rdur := strings.Split(scope.LeaseDuration, ":")
	// dur := time.Duration(0)
	// // days
	// day, err := strconv.Atoi(rdur[0])
	// if err != nil {
	// 	log.Println(err)
	// 	continue
	// }
	// dur += day * 24 * time.Hour
	kscope := dhcpv1.Scope{
		ObjectMeta: v1.ObjectMeta{
			Name: sc.Name,
		},
		TypeMeta: v1.TypeMeta{
			Kind:       "Scope",
			APIVersion: "dhcp.beryju.org/v1",
		},
		Spec: dhcpv1.ScopeSpec{
			SubnetCIDR: cidr.String(),
			Range: dhcpv1.ScopeSpecRange{
				Start: sc.StartRange,
				End:   sc.EndRange,
			},
			LeaseTemplate: &dhcpv1.LeaseCommonSpec{
				OptionSet: corev1.LocalObjectReference{
					Name: sc.Name,
				},
			},
		},
	}
	c.writeFile(&kscope)
	c.convertScopeOptionValues(kscope, sc.OptionValues.OptionValue)
	for _, res := range sc.Reservations.Reservation {
		c.convertReservation(kscope, res)
	}
	for _, l := range sc.Leases.Lease {
		c.convertLease(kscope, l)
	}
}

func (c *Converter) convertScopeOptionValues(ks dhcpv1.Scope, val []OptionValue) {
	kos := dhcpv1.OptionSet{
		ObjectMeta: v1.ObjectMeta{
			Name: ks.Name,
		},
		TypeMeta: v1.TypeMeta{
			Kind:       "OptionSet",
			APIVersion: "dhcp.beryju.org/v1",
		},
		Spec: dhcpv1.OptionSetSpec{},
	}
	for _, optv := range val {
		tag, err := strconv.Atoi(optv.OptionId)
		if err != nil {
			c.l.Error(err, "failed to convert optionID to int")
			continue
		}
		t := uint8(tag)
		v := optv.Value[0]
		kopt := dhcpv1.Option{
			Tag:   &t,
			Value: &v,
		}
		kos.Spec.Options = append(kos.Spec.Options, &kopt)
	}
	c.writeFile(&kos)
}

func (c *Converter) getIdentifier(clientId string) string {
	if strings.Count(clientId, "-") == 5 {
		return strings.ReplaceAll(clientId, "-", ":")
	}
	return strings.ReplaceAll(clientId, "-", "")
}

func (c *Converter) convertReservation(ks dhcpv1.Scope, r Reservation) {
	name := slug.Make(fmt.Sprintf("%s-%s", ks.Name, r.Name))
	lease := dhcpv1.Lease{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		TypeMeta: v1.TypeMeta{
			Kind:       "Lease",
			APIVersion: "dhcp.beryju.org/v1",
		},
		Spec: dhcpv1.LeaseSpec{
			Identifier: c.getIdentifier(r.ClientId),
			Hostname:   r.Name,
			Address:    r.IPAddress,
			Scope: corev1.LocalObjectReference{
				Name: ks.Name,
			},
			LeaseCommonSpec: dhcpv1.LeaseCommonSpec{
				AddressLeaseTime: "-1",
			},
		},
	}
	c.writeFile(&lease)
}

func (c *Converter) convertLease(ks dhcpv1.Scope, l Lease) {
	if l.HostName == "BAD_ADDRESS" {
		return
	}
	name := slug.Make(fmt.Sprintf("%s-%s", ks.Name, l.HostName))
	lease := dhcpv1.Lease{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		TypeMeta: v1.TypeMeta{
			Kind:       "Lease",
			APIVersion: "dhcp.beryju.org/v1",
		},
		Spec: dhcpv1.LeaseSpec{
			Identifier: c.getIdentifier(l.ClientId),
			Address:    l.IPAddress,
			Hostname:   l.HostName,
			Scope: corev1.LocalObjectReference{
				Name: ks.Name,
			},
		},
	}
	c.writeFile(&lease)
}

func (c *Converter) writeFile(o client.Object) {
	y, err := json.Marshal(o)
	if err != nil {
		log.Println(err)
	}
	path := fmt.Sprintf("./%s/%s-%s.json", c.out, o.GetObjectKind().GroupVersionKind().Kind, o.GetName())
	err = ioutil.WriteFile(path, y, os.ModeSticky|os.ModePerm)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("Successfully wrote %s\n", path)
	}
}
