package twointerior

import (
	"fmt"
	"github.com/fgiorgetti/qpid-dispatch-go-tests/pkg/api/client/amqp"
	"github.com/fgiorgetti/qpid-dispatch-go-tests/pkg/api/client/amqp/qeclients"
	"github.com/fgiorgetti/qpid-dispatch-go-tests/pkg/framework"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"strconv"
)

const (
	numSenders   = 2
	numReceivers = 2
	messageCount = 1000
	messageBody  = "1234567890123456789012345678901234567890123456789012345678901234"
	address      = "anycast.sample"
)

var _ = ginkgo.Describe("TwoInteriorMessaging", func() {

	var (
		ctx1          *framework.ContextData
		ctx2          *framework.ContextData
		sendersCtx1   []amqp.Client
		receiversCtx1 []amqp.Client
		sendersCtx2   []amqp.Client
		receiversCtx2 []amqp.Client
	)

	// Initialize after frameworks have been created
	ginkgo.JustBeforeEach(func() {
		ctx1 = FrameworkQdrOne.GetFirstContext()
		ctx2 = FrameworkQdrTwo.GetFirstContext()
		sendersCtx1 = createSenders(*ctx1, address, QdrOneName, numSenders)
		receiversCtx1 = createReceivers(*ctx1, address, QdrOneName, numReceivers)
		sendersCtx2 = createSenders(*ctx2, address, QdrTwoName, numSenders)
		receiversCtx2 = createReceivers(*ctx2, address, QdrTwoName, numReceivers)
	})

	ginkgo.It("Exchanges anycast messages using multiple clients against the distributed mesh", func() {
		ginkgo.By("Creating 2 receivers on each cluster")
		// Starting receivers
		startClients(receiversCtx1)
		startClients(receiversCtx2)
		ginkgo.By("Creating 2 senders on each cluster")
		// Starting senders
		startClients(sendersCtx1)
		startClients(sendersCtx2)

		// Wait all clients to complete
		waitClients(receiversCtx1, receiversCtx2, sendersCtx1, sendersCtx2)

		// Validating status and number of processed messages
		ginkgo.By("Validating all receivers completed and all messages exchanged")
		validateResults(messageCount, false, receiversCtx1, receiversCtx2)
		ginkgo.By("Validating all senders completed and all messages exchanged")
		validateResults(messageCount, true, sendersCtx1, sendersCtx2)
	})
})

// waitClients iterates through all slices and wais for all cliens to finish
func waitClients(clientSlices ...[]amqp.Client) {
	for _, clients := range clientSlices {
		for _, c := range clients {
			c.Wait()
		}
	}
}

// validateResults simply validates if clients completed successfully
// and result data contains expected amount of messages.
func validateResults(messageCount int, greaterEqual bool, clientSlices ...[]amqp.Client) {
	for _, clients := range clientSlices {
		for _, c := range clients {
			gomega.Expect(c.Status()).To(gomega.Equal(amqp.Success))
			// Needed cause Python sender currently does not output number of released or modified messages
			if greaterEqual {
				gomega.Expect(c.Result().Delivered).To(gomega.BeNumerically(">=", messageCount))
			} else {
				gomega.Expect(c.Result().Delivered).To(gomega.Equal(messageCount))
			}
		}
	}
}

// startClients Deploys all clients from provided slice. Clients must be prepared already.
func startClients(clients []amqp.Client) {
	for _, c := range clients {
		err := c.Deploy()
		gomega.Expect(err).To(gomega.BeNil())
	}
}

// createReceivers returns a slice of receiver clients ready to be deployed
func createReceivers(context framework.ContextData, address string, deploymentName string, number int) []amqp.Client {
	var receivers []amqp.Client
	receivers = make([]amqp.Client, 0)
	for i := 0; i < number; i++ {
		builder := qeclients.NewReceiverBuilder(qeclients.Python)
		url := getURL(address, deploymentName, context)
		builder.New("receiver"+strconv.Itoa(i), context, url)
		builder.Timeout(amqp.TimeoutDefaultSecs)
		builder.Messages(messageCount)
		receiver, err := builder.Build()
		gomega.Expect(err).To(gomega.BeNil())
		// Appending the receiver
		receivers = append(receivers, receiver)
	}
	return receivers
}

// createSenders returns a slice of sender clients ready to be deployed
func createSenders(context framework.ContextData, address string, deploymentName string, number int) []amqp.Client {
	var senders []amqp.Client
	senders = make([]amqp.Client, 0)
	for i := 0; i < number; i++ {
		// Create a builder for the sender
		builder := qeclients.NewSenderBuilder(qeclients.Python)
		url := getURL(address, deploymentName, context)
		builder.New("sender-"+strconv.Itoa(i+1), context, url)
		builder.Timeout(amqp.TimeoutDefaultSecs)
		builder.Messages(messageCount)
		builder.MessageContent(messageBody)
		sender, err := builder.Build()
		gomega.Expect(err).To(gomega.BeNil())
		// Appending the sender
		senders = append(senders, sender)
	}
	return senders
}

// getURL returns an amqp url based on provided deployment name and context
func getURL(address string, deploymentName string, context framework.ContextData) string {
	return fmt.Sprintf("amqp://%s.%s.svc.cluster.local/%s", deploymentName, context.Namespace, address)
}
