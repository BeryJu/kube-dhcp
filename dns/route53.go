package dns

import (
	"errors"
	"strconv"

	dhcpv1 "beryju.org/kube-dhcp/api/v1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	r53 "github.com/aws/aws-sdk-go/service/route53"
)

type Route53DNSProvider struct {
	r             *r53.Route53
	zoneId        string
	reverseZoneId string
	ttl           int64
	z             *r53.GetHostedZoneOutput
}

func NewRoute53Provider(config map[string]string) (*Route53DNSProvider, error) {
	p := &Route53DNSProvider{}

	if zone, zoneOk := config["zoneId"]; !zoneOk {
		return nil, errors.New("zoneId not set in scope config")
	} else {
		p.zoneId = zone
	}

	if reverseZoneId, reverseZoneIdOk := config["reverseZoneIdId"]; reverseZoneIdOk {
		p.reverseZoneId = reverseZoneId
	}

	ttl, err := strconv.Atoi(config["ttl"])
	if err != nil {
		ttl = 3600
	}
	p.ttl = int64(ttl)

	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	p.r = route53.New(sess)

	zone, err := p.r.GetHostedZone(&r53.GetHostedZoneInput{
		Id: &p.zoneId,
	})
	if err != nil {
		return nil, err
	}
	p.z = zone

	return p, nil
}

func (r *Route53DNSProvider) updateRecord(chg *route53.Change, zone string) error {
	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				chg,
			},
		},
		HostedZoneId: aws.String(zone),
	}
	_, err := r.r.ChangeResourceRecordSets(params)
	if err != nil {
		return err
	}
	return nil

}

func (r *Route53DNSProvider) CreateRecord(lease *dhcpv1.Lease) error {
	fwd := aws.String(lease.Spec.Hostname + "." + *r.z.HostedZone.Name)
	err := r.updateRecord(&route53.Change{
		Action: aws.String(r53.ChangeActionUpsert),
		ResourceRecordSet: &route53.ResourceRecordSet{
			Name: fwd,
			Type: aws.String(r53.RRTypeA),
			ResourceRecords: []*route53.ResourceRecord{
				{
					Value: aws.String(lease.Spec.Address),
				},
			},
			TTL: aws.Int64(r.ttl),
		},
	}, r.zoneId)
	if err != nil {
		return err
	}
	err = r.updateRecord(&route53.Change{
		Action: aws.String(r53.ChangeActionUpsert),
		ResourceRecordSet: &route53.ResourceRecordSet{
			Name: aws.String(reverseDNSRecord(lease.Spec.Address)),
			Type: aws.String(r53.RRTypePtr),
			ResourceRecords: []*route53.ResourceRecord{
				{
					Value: fwd,
				},
			},
			TTL: aws.Int64(r.ttl),
		},
	}, r.reverseZoneId)
	if err != nil {
		return err
	}
	return nil
}

func (r *Route53DNSProvider) UpdateRecord(lease *dhcpv1.Lease) error {
	return nil
}

func (r *Route53DNSProvider) DeleteRecord(lease *dhcpv1.Lease) error {
	fwd := aws.String(lease.Spec.Hostname + "." + *r.z.HostedZone.Name)
	err := r.updateRecord(&route53.Change{
		Action: aws.String(r53.ChangeActionDelete),
		ResourceRecordSet: &route53.ResourceRecordSet{
			Name: fwd,
			Type: aws.String(r53.RRTypeA),
			ResourceRecords: []*route53.ResourceRecord{
				{
					Value: aws.String(lease.Spec.Address),
				},
			},
			TTL: aws.Int64(r.ttl),
		},
	}, r.zoneId)
	if err != nil {
		return err
	}
	err = r.updateRecord(&route53.Change{
		Action: aws.String(r53.ChangeActionDelete),
		ResourceRecordSet: &route53.ResourceRecordSet{
			Name: aws.String(reverseDNSRecord(lease.Spec.Address)),
			Type: aws.String(r53.RRTypePtr),
			ResourceRecords: []*route53.ResourceRecord{
				{
					Value: fwd,
				},
			},
			TTL: aws.Int64(r.ttl),
		},
	}, r.reverseZoneId)
	if err != nil {
		return err
	}
	return nil
}
