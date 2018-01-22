package azure_test

import (
	"errors"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevievelesperance/leftovers/azure"
	"github.com/genevievelesperance/leftovers/azure/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Groups", func() {
	var (
		client *fakes.GroupsClient
		logger *fakes.Logger
		filter string

		groups azure.Groups
	)

	BeforeEach(func() {
		client = &fakes.GroupsClient{}
		logger = &fakes.Logger{}
		filter = "banana"

		groups = azure.NewGroups(client, logger)
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			logger.PromptCall.Returns.Proceed = true
			client.ListCall.Returns.Output = resources.GroupListResult{
				Value: &[]resources.Group{{
					Name: aws.String("banana-group"),
				}},
			}
			errChan := make(chan error, 1)
			errChan <- nil
			client.DeleteCall.Returns.Error = errChan
		})

		It("deletes resource groups", func() {
			err := groups.Delete(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListCall.CallCount).To(Equal(1))
			Expect(client.DeleteCall.CallCount).To(Equal(1))
			Expect(client.DeleteCall.Receives.Name).To(Equal("banana-group"))
			Expect(logger.PrintfCall.Messages).To(Equal([]string{"SUCCESS deleting resource group banana-group\n"}))
		})

		Context("when client fails to list resource groups", func() {
			BeforeEach(func() {
				client.ListCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				err := groups.Delete(filter)
				Expect(err).To(MatchError("Listing resource groups: some error"))

				Expect(client.ListCall.CallCount).To(Equal(1))
				Expect(client.DeleteCall.CallCount).To(Equal(0))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptCall.Returns.Proceed = false
			})

			It("does not delete the resource group", func() {
				err := groups.Delete(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete resource group banana-group?"))
				Expect(client.DeleteCall.CallCount).To(Equal(0))
			})
		})

		Context("when the resource group name does not contain the filter", func() {
			It("does not delete it", func() {
				err := groups.Delete("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.CallCount).To(Equal(0))
				Expect(client.DeleteCall.CallCount).To(Equal(0))
			})
		})

		Context("when client fails to delete the resource group", func() {
			BeforeEach(func() {
				client.ListCall.Returns.Output = resources.GroupListResult{
					Value: &[]resources.Group{{
						Name: aws.String("banana-group"),
					}},
				}
				errChan := make(chan error, 1)
				errChan <- errors.New("some error")
				client.DeleteCall.Returns.Error = errChan
			})

			It("logs the error", func() {
				err := groups.Delete(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(client.ListCall.CallCount).To(Equal(1))
				Expect(client.DeleteCall.CallCount).To(Equal(1))
				Expect(client.DeleteCall.Receives.Name).To(Equal("banana-group"))
				Expect(logger.PrintfCall.Messages).To(Equal([]string{"ERROR deleting resource group banana-group: some error\n"}))
			})
		})
	})
})
