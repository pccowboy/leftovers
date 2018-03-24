package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/ec2/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SecurityGroups", func() {
	var (
		client *fakes.SecurityGroupsClient
		logger *fakes.Logger

		securityGroups ec2.SecurityGroups
	)

	BeforeEach(func() {
		client = &fakes.SecurityGroupsClient{}
		logger = &fakes.Logger{}
		logger.PromptWithDetailsCall.Returns.Proceed = true

		securityGroups = ec2.NewSecurityGroups(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			client.DescribeSecurityGroupsCall.Returns.Output = &awsec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []*awsec2.SecurityGroup{{
					GroupName: aws.String("banana-group"),
					GroupId:   aws.String("the-group-id"),
				}},
			}
			filter = "banana"
		})

		It("deletes ec2 security groups", func() {
			items, err := securityGroups.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeSecurityGroupsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("security group"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-group"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the client fails to describe security groups", func() {
			BeforeEach(func() {
				client.DescribeSecurityGroupsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := securityGroups.List(filter)
				Expect(err).To(MatchError("Describing security groups: some error"))
			})
		})

		Context("when the security group name does not contain the filter", func() {
			It("does not try deleting them", func() {
				items, err := securityGroups.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DescribeSecurityGroupsCall.CallCount).To(Equal(1))
				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the security group has tags", func() {
			BeforeEach(func() {
				client.DescribeSecurityGroupsCall.Returns.Output = &awsec2.DescribeSecurityGroupsOutput{
					SecurityGroups: []*awsec2.SecurityGroup{{
						GroupName: aws.String("banana-group"),
						GroupId:   aws.String("the-group-id"),
						Tags:      []*awsec2.Tag{{Key: aws.String("the-key"), Value: aws.String("the-value")}},
					}},
				}
			})

			It("uses the tag in the prompt", func() {
				_, err := securityGroups.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-group (the-key:the-value)"))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not delete the security group", func() {
				items, err := securityGroups.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
