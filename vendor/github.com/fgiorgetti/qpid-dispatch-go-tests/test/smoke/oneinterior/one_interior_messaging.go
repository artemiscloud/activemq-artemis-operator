package oneinterior

import (
	"fmt"
	"github.com/fgiorgetti/qpid-dispatch-go-tests/pkg/api/client/amqp"
	"github.com/fgiorgetti/qpid-dispatch-go-tests/pkg/api/client/amqp/qeclients"
	"github.com/fgiorgetti/qpid-dispatch-go-tests/pkg/framework"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("OneInteriorMessaging", func() {
	// Clients to be initialized for each test
	var (
		ctx      *framework.ContextData
		sender   amqp.Client
		receiver amqp.Client
		url      string
		err      error
	)

	const (
		// The message body to be used
		MessageBody = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		// Amount of messages to exchange
		MessageCount = 100
	)

	// Run the clients
	ginkgo.JustBeforeEach(func() {
		ctx = Framework.GetFirstContext()

		// Url to use on both clients
		url = fmt.Sprintf("amqp://%s:5672/one_interior_messaging", DeployName)

		// Building sender client
		sender, err = qeclients.NewAmqpSender(qeclients.Python, "sender", *ctx, url, MessageCount, MessageBody)
		gomega.Expect(err).To(gomega.BeNil())

		// Building receiver client
		receiver, err = qeclients.NewAmqpReceiver(qeclients.Python, "receiver", *ctx, url, MessageCount)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("Exchange messages through router mesh", func() {
		// Deploying clients
		err = sender.Deploy()
		gomega.Expect(err).To(gomega.BeNil())
		err = receiver.Deploy()
		gomega.Expect(err).To(gomega.BeNil())

		// Waiting till client completes (or is interrupted)
		sender.Wait()
		receiver.Wait()

		// Validating results
		senderResult := sender.Result()
		receiverResult := receiver.Result()

		// Ensure results obtained
		gomega.Expect(senderResult).NotTo(gomega.BeNil())
		gomega.Expect(receiverResult).NotTo(gomega.BeNil())

		// Validate sent/received messages
		gomega.Expect(senderResult.Delivered).To(gomega.Equal(MessageCount))
		gomega.Expect(receiverResult.Delivered).To(gomega.Equal(MessageCount))
	})
})
