package route53

import (
	"strconv"

	dhcpv1 "beryju.org/kube-dhcp/api/v1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	r53 "github.com/aws/aws-sdk-go/service/route53"
)

type Route53DNSProvider struct {
	r      *r53.Route53
	zoneId string
	ttl    int64
	z      *r53.GetHostedZoneOutput
}

func New(config map[string]string) *Route53DNSProvider {
	p := &Route53DNSProvider{}
	// TODO: error handling
	p.zoneId = config["zoneId"]
	ttl, err := strconv.Atoi(config["ttl"])
	if err != nil {
		ttl = 3600
	}
	p.ttl = int64(ttl)

	sess, err := session.NewSession()
	if err != nil {
		// todo: error handling
		panic(err)
	}

	p.r = route53.New(sess)

	zone, err := p.r.GetHostedZone(&r53.GetHostedZoneInput{
		Id: &p.zoneId,
	})
	if err != nil {
		panic(err)
	}
	p.z = zone

	return p
}

func (r *Route53DNSProvider) CreateRecord(lease *dhcpv1.Lease) error {
	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(r53.ChangeActionUpsert),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(lease.Spec.Hostname + "." + *r.z.HostedZone.Name),
						Type: aws.String(r53.RRTypeA),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(lease.Spec.Address),
							},
						},
						TTL: aws.Int64(r.ttl),
					},
				},
			},
		},
		HostedZoneId: aws.String(r.zoneId),
	}
	_, err := r.r.ChangeResourceRecordSets(params)
	if err != nil {
		return err
	}
	return nil
}

func (r *Route53DNSProvider) UpdateRecord(lease *dhcpv1.Lease) error {
	return nil
}

func (r *Route53DNSProvider) DeleteRecord(lease *dhcpv1.Lease) error {
	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(r53.ChangeActionDelete),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(lease.Spec.Hostname + "." + *r.z.HostedZone.Name),
						Type: aws.String(r53.RRTypeA),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(lease.Spec.Address),
							},
						},
						TTL: aws.Int64(r.ttl),
					},
				},
			},
		},
		HostedZoneId: aws.String(r.zoneId),
	}
	_, err := r.r.ChangeResourceRecordSets(params)
	if err != nil {
		return err
	}
	return nil
}
