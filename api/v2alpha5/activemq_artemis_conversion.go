package v2alpha5

import (
	"github.com/RHsyseng/operator-utils/pkg/olm"
	v1beta1 "github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("v2alpha5Conversion")

// ConvertTo converts this v2alpha5 to v1beta1. (upgrade)
func (src *ActiveMQArtemis) ConvertTo(dstRaw conversion.Hub) error {
	log.Info("converting v2alpha5 to v1beta1")
	dst := dstRaw.(*v1beta1.ActiveMQArtemis)

	// Meta
	dst.ObjectMeta = src.ObjectMeta

	// Status
	dst.Status.PodStatus = olm.DeploymentStatus{
		Ready:    src.Status.PodStatus.Ready,
		Starting: src.Status.PodStatus.Starting,
		Stopped:  src.Status.PodStatus.Stopped,
	}

	// Spec
	dst.Spec.AdminUser = src.Spec.AdminUser
	dst.Spec.AdminPassword = src.Spec.AdminPassword

	dst.Spec.DeploymentPlan.Image = src.Spec.DeploymentPlan.Image
	dst.Spec.DeploymentPlan.InitImage = src.Spec.DeploymentPlan.InitImage
	dst.Spec.DeploymentPlan.Size = src.Spec.DeploymentPlan.Size
	dst.Spec.DeploymentPlan.RequireLogin = src.Spec.DeploymentPlan.RequireLogin
	dst.Spec.DeploymentPlan.PersistenceEnabled = src.Spec.DeploymentPlan.PersistenceEnabled
	dst.Spec.DeploymentPlan.JournalType = src.Spec.DeploymentPlan.JournalType
	dst.Spec.DeploymentPlan.MessageMigration = src.Spec.DeploymentPlan.MessageMigration
	dst.Spec.DeploymentPlan.Resources = *src.Spec.DeploymentPlan.Resources.DeepCopy()
	dst.Spec.DeploymentPlan.Storage.Size = src.Spec.DeploymentPlan.Storage.Size
	dst.Spec.DeploymentPlan.JolokiaAgentEnabled = src.Spec.DeploymentPlan.JolokiaAgentEnabled
	dst.Spec.DeploymentPlan.ManagementRBACEnabled = src.Spec.DeploymentPlan.ManagementRBACEnabled
	dst.Spec.DeploymentPlan.ExtraMounts.ConfigMaps = src.Spec.DeploymentPlan.ExtraMounts.ConfigMaps
	dst.Spec.DeploymentPlan.ExtraMounts.Secrets = src.Spec.DeploymentPlan.ExtraMounts.Secrets
	dst.Spec.DeploymentPlan.Clustered = src.Spec.DeploymentPlan.Clustered
	dst.Spec.DeploymentPlan.PodSecurity.RunAsUser = src.Spec.DeploymentPlan.PodSecurity.RunAsUser
	dst.Spec.DeploymentPlan.PodSecurity.ServiceAccountName = src.Spec.DeploymentPlan.PodSecurity.ServiceAccountName
	if src.Spec.DeploymentPlan.LivenessProbe.TimeoutSeconds != nil {
		dst.Spec.DeploymentPlan.LivenessProbe = &corev1.Probe{
			TimeoutSeconds: *src.Spec.DeploymentPlan.LivenessProbe.TimeoutSeconds,
		}
	}
	if src.Spec.DeploymentPlan.ReadinessProbe.TimeoutSeconds != nil {
		dst.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			TimeoutSeconds: *src.Spec.DeploymentPlan.ReadinessProbe.TimeoutSeconds,
		}
	}
	src.Spec.DeploymentPlan.EnableMetricsPlugin = dst.Spec.DeploymentPlan.EnableMetricsPlugin

	dst.Spec.Acceptors = make([]v1beta1.AcceptorType, len(src.Spec.Acceptors))
	for i, acc := range src.Spec.Acceptors {
		dst.Spec.Acceptors[i] = v1beta1.AcceptorType{
			Name:                              acc.Name,
			Port:                              acc.Port,
			Protocols:                         acc.Protocols,
			SSLEnabled:                        acc.SSLEnabled,
			SSLSecret:                         acc.SSLSecret,
			EnabledCipherSuites:               acc.EnabledCipherSuites,
			EnabledProtocols:                  acc.EnabledProtocols,
			NeedClientAuth:                    acc.NeedClientAuth,
			WantClientAuth:                    acc.WantClientAuth,
			VerifyHost:                        acc.VerifyHost,
			SSLProvider:                       acc.SSLProvider,
			SNIHost:                           acc.SNIHost,
			Expose:                            acc.Expose,
			AnycastPrefix:                     acc.AnycastPrefix,
			MulticastPrefix:                   acc.MulticastPrefix,
			ConnectionsAllowed:                acc.ConnectionsAllowed,
			AMQPMinLargeMessageSize:           acc.AMQPMinLargeMessageSize,
			SupportAdvisory:                   acc.SupportAdvisory,
			SuppressInternalManagementObjects: acc.SuppressInternalManagementObjects,
		}
	}

	dst.Spec.Connectors = make([]v1beta1.ConnectorType, len(src.Spec.Connectors))
	for i, conn := range src.Spec.Connectors {
		dst.Spec.Connectors[i] = v1beta1.ConnectorType{
			Name:                conn.Name,
			Type:                conn.Type,
			Host:                conn.Host,
			Port:                conn.Port,
			SSLEnabled:          conn.SSLEnabled,
			SSLSecret:           conn.SSLSecret,
			EnabledCipherSuites: conn.EnabledCipherSuites,
			EnabledProtocols:    conn.EnabledProtocols,
			NeedClientAuth:      conn.NeedClientAuth,
			WantClientAuth:      conn.WantClientAuth,
			VerifyHost:          conn.VerifyHost,
			SSLProvider:         conn.SSLProvider,
			SNIHost:             conn.SNIHost,
			Expose:              conn.Expose,
		}
	}

	dst.Spec.Console = v1beta1.ConsoleType{
		Expose:        src.Spec.Console.Expose,
		SSLEnabled:    src.Spec.Console.SSLEnabled,
		SSLSecret:     src.Spec.Console.SSLSecret,
		UseClientAuth: src.Spec.Console.UseClientAuth,
	}

	dst.Spec.Version = src.Spec.Version

	dst.Spec.Upgrades.Enabled = src.Spec.Upgrades.Enabled
	dst.Spec.Upgrades.Minor = src.Spec.Upgrades.Minor

	dst.Spec.AddressSettings.ApplyRule = src.Spec.AddressSettings.ApplyRule
	dst.Spec.AddressSettings.AddressSetting = make([]v1beta1.AddressSettingType, len(src.Spec.AddressSettings.AddressSetting))
	for i, add := range src.Spec.AddressSettings.AddressSetting {
		dst.Spec.AddressSettings.AddressSetting[i] = v1beta1.AddressSettingType{
			DeadLetterAddress:                    add.DeadLetterAddress,
			AutoCreateDeadLetterResources:        add.AutoCreateDeadLetterResources,
			DeadLetterQueuePrefix:                add.DeadLetterQueuePrefix,
			DeadLetterQueueSuffix:                add.DeadLetterQueueSuffix,
			ExpiryAddress:                        add.ExpiryAddress,
			AutoCreateExpiryResources:            add.AutoCreateDeadLetterResources,
			ExpiryQueuePrefix:                    add.ExpiryQueuePrefix,
			ExpiryQueueSuffix:                    add.ExpiryQueueSuffix,
			ExpiryDelay:                          add.ExpiryDelay,
			MinExpiryDelay:                       add.MinExpiryDelay,
			MaxExpiryDelay:                       add.MaxExpiryDelay,
			RedeliveryDelay:                      add.RedeliveryDelay,
			MaxRedeliveryDelay:                   add.MaxRedeliveryDelay,
			MaxDeliveryAttempts:                  add.MaxDeliveryAttempts,
			MaxSizeBytes:                         add.MaxSizeBytes,
			MaxSizeBytesRejectThreshold:          add.MaxSizeBytesRejectThreshold,
			PageSizeBytes:                        add.PageSizeBytes,
			PageMaxCacheSize:                     add.PageMaxCacheSize,
			AddressFullPolicy:                    add.AddressFullPolicy,
			MessageCounterHistoryDayLimit:        add.MessageCounterHistoryDayLimit,
			LastValueQueue:                       add.LastValueQueue,
			DefaultLastValueQueue:                add.DefaultLastValueQueue,
			DefaultLastValueKey:                  add.DefaultLastValueKey,
			DefaultNonDestructive:                add.DefaultNonDestructive,
			DefaultExclusiveQueue:                add.DefaultExclusiveQueue,
			DefaultGroupRebalance:                add.DefaultGroupRebalance,
			DefaultGroupRebalancePauseDispatch:   add.DefaultGroupRebalancePauseDispatch,
			DefaultGroupBuckets:                  add.DefaultGroupBuckets,
			DefaultGroupFirstKey:                 add.DefaultGroupFirstKey,
			DefaultConsumersBeforeDispatch:       add.DefaultConsumersBeforeDispatch,
			DefaultDelayBeforeDispatch:           add.DefaultDelayBeforeDispatch,
			RedistributionDelay:                  add.RedistributionDelay,
			SendToDlaOnNoRoute:                   add.SendToDlaOnNoRoute,
			SlowConsumerThreshold:                add.SlowConsumerThreshold,
			SlowConsumerPolicy:                   add.SlowConsumerPolicy,
			SlowConsumerCheckPeriod:              add.SlowConsumerCheckPeriod,
			AutoCreateJmsQueues:                  add.AutoCreateJmsQueues,
			AutoDeleteJmsQueues:                  add.AutoDeleteJmsQueues,
			AutoCreateJmsTopics:                  add.AutoCreateJmsTopics,
			AutoDeleteJmsTopics:                  add.AutoDeleteJmsTopics,
			AutoCreateQueues:                     add.AutoCreateQueues,
			AutoDeleteQueues:                     add.AutoDeleteQueues,
			AutoDeleteCreatedQueues:              add.AutoDeleteCreatedQueues,
			AutoDeleteQueuesDelay:                add.AutoDeleteQueuesDelay,
			AutoDeleteQueuesMessageCount:         add.AutoDeleteQueuesMessageCount,
			ConfigDeleteQueues:                   add.ConfigDeleteQueues,
			AutoCreateAddresses:                  add.AutoCreateAddresses,
			AutoDeleteAddresses:                  add.AutoDeleteAddresses,
			AutoDeleteAddressesDelay:             add.AutoDeleteAddressesDelay,
			ConfigDeleteAddresses:                add.ConfigDeleteAddresses,
			ManagementBrowsePageSize:             add.ManagementBrowsePageSize,
			DefaultPurgeOnNoConsumers:            add.DefaultPurgeOnNoConsumers,
			DefaultMaxConsumers:                  add.DefaultMaxConsumers,
			DefaultQueueRoutingType:              add.DefaultQueueRoutingType,
			DefaultAddressRoutingType:            add.DefaultAddressRoutingType,
			DefaultConsumerWindowSize:            add.DefaultConsumerWindowSize,
			DefaultRingSize:                      add.DefaultRingSize,
			RetroactiveMessageCount:              add.RetroactiveMessageCount,
			EnableMetrics:                        add.EnableMetrics,
			Match:                                add.Match,
			ManagementMessageAttributeSizeLimit:  add.ManagementMessageAttributeSizeLimit,
			SlowConsumerThresholdMeasurementUnit: add.SlowConsumerThresholdMeasurementUnit,
			EnableIngressTimestamp:               add.EnableIngressTimestamp,
		}
	}

	log.Info("converted v2alpha5 to v1beta1")
	return nil
}

// ConvertFrom converts from the Hub version (v1beta1) to v2alpha5. (downgrade)
func (dst *ActiveMQArtemis) ConvertFrom(srcRaw conversion.Hub) error {
	log.Info("converting from v1beta1 to v2alpha5")
	src := srcRaw.(*v1beta1.ActiveMQArtemis)

	// Meta
	dst.ObjectMeta = src.ObjectMeta

	// Status
	dst.Status.PodStatus = olm.DeploymentStatus{
		Ready:    src.Status.PodStatus.Ready,
		Starting: src.Status.PodStatus.Starting,
		Stopped:  src.Status.PodStatus.Stopped,
	}

	// Spec
	dst.Spec.AdminUser = src.Spec.AdminUser
	dst.Spec.AdminPassword = src.Spec.AdminPassword

	dst.Spec.DeploymentPlan.Image = src.Spec.DeploymentPlan.Image
	dst.Spec.DeploymentPlan.InitImage = src.Spec.DeploymentPlan.InitImage
	dst.Spec.DeploymentPlan.Size = src.Spec.DeploymentPlan.Size
	dst.Spec.DeploymentPlan.RequireLogin = src.Spec.DeploymentPlan.RequireLogin
	dst.Spec.DeploymentPlan.PersistenceEnabled = src.Spec.DeploymentPlan.PersistenceEnabled
	dst.Spec.DeploymentPlan.JournalType = src.Spec.DeploymentPlan.JournalType
	dst.Spec.DeploymentPlan.MessageMigration = src.Spec.DeploymentPlan.MessageMigration
	dst.Spec.DeploymentPlan.Resources = *src.Spec.DeploymentPlan.Resources.DeepCopy()
	dst.Spec.DeploymentPlan.Storage.Size = src.Spec.DeploymentPlan.Storage.Size
	dst.Spec.DeploymentPlan.JolokiaAgentEnabled = src.Spec.DeploymentPlan.JolokiaAgentEnabled
	dst.Spec.DeploymentPlan.ManagementRBACEnabled = src.Spec.DeploymentPlan.ManagementRBACEnabled
	dst.Spec.DeploymentPlan.ExtraMounts.ConfigMaps = src.Spec.DeploymentPlan.ExtraMounts.ConfigMaps
	dst.Spec.DeploymentPlan.ExtraMounts.Secrets = src.Spec.DeploymentPlan.ExtraMounts.Secrets
	dst.Spec.DeploymentPlan.Clustered = src.Spec.DeploymentPlan.Clustered
	dst.Spec.DeploymentPlan.PodSecurity.RunAsUser = src.Spec.DeploymentPlan.PodSecurity.RunAsUser
	dst.Spec.DeploymentPlan.PodSecurity.ServiceAccountName = src.Spec.DeploymentPlan.PodSecurity.ServiceAccountName
	if src.Spec.DeploymentPlan.LivenessProbe != nil {
		dst.Spec.DeploymentPlan.LivenessProbe = LivenessProbeType{
			TimeoutSeconds: &src.Spec.DeploymentPlan.LivenessProbe.TimeoutSeconds,
		}
	}
	if src.Spec.DeploymentPlan.ReadinessProbe != nil {
		dst.Spec.DeploymentPlan.ReadinessProbe = ReadinessProbeType{
			TimeoutSeconds: &src.Spec.DeploymentPlan.ReadinessProbe.TimeoutSeconds,
		}
	}
	src.Spec.DeploymentPlan.EnableMetricsPlugin = dst.Spec.DeploymentPlan.EnableMetricsPlugin

	dst.Spec.Acceptors = make([]AcceptorType, len(src.Spec.Acceptors))
	for i, acc := range src.Spec.Acceptors {
		dst.Spec.Acceptors[i] = AcceptorType{
			Name:                              acc.Name,
			Port:                              acc.Port,
			Protocols:                         acc.Protocols,
			SSLEnabled:                        acc.SSLEnabled,
			SSLSecret:                         acc.SSLSecret,
			EnabledCipherSuites:               acc.EnabledCipherSuites,
			EnabledProtocols:                  acc.EnabledProtocols,
			NeedClientAuth:                    acc.NeedClientAuth,
			WantClientAuth:                    acc.WantClientAuth,
			VerifyHost:                        acc.VerifyHost,
			SSLProvider:                       acc.SSLProvider,
			SNIHost:                           acc.SNIHost,
			Expose:                            acc.Expose,
			AnycastPrefix:                     acc.AnycastPrefix,
			MulticastPrefix:                   acc.MulticastPrefix,
			ConnectionsAllowed:                acc.ConnectionsAllowed,
			AMQPMinLargeMessageSize:           acc.AMQPMinLargeMessageSize,
			SupportAdvisory:                   acc.SupportAdvisory,
			SuppressInternalManagementObjects: acc.SuppressInternalManagementObjects,
		}
	}

	dst.Spec.Connectors = make([]ConnectorType, len(src.Spec.Connectors))
	for i, conn := range src.Spec.Connectors {
		dst.Spec.Connectors[i] = ConnectorType{
			Name:                conn.Name,
			Type:                conn.Type,
			Host:                conn.Host,
			Port:                conn.Port,
			SSLEnabled:          conn.SSLEnabled,
			SSLSecret:           conn.SSLSecret,
			EnabledCipherSuites: conn.EnabledCipherSuites,
			EnabledProtocols:    conn.EnabledProtocols,
			NeedClientAuth:      conn.NeedClientAuth,
			WantClientAuth:      conn.WantClientAuth,
			VerifyHost:          conn.VerifyHost,
			SSLProvider:         conn.SSLProvider,
			SNIHost:             conn.SNIHost,
			Expose:              conn.Expose,
		}
	}

	dst.Spec.Console = ConsoleType{
		Expose:        src.Spec.Console.Expose,
		SSLEnabled:    src.Spec.Console.SSLEnabled,
		SSLSecret:     src.Spec.Console.SSLSecret,
		UseClientAuth: src.Spec.Console.UseClientAuth,
	}

	dst.Spec.Version = src.Spec.Version

	dst.Spec.Upgrades.Enabled = src.Spec.Upgrades.Enabled
	dst.Spec.Upgrades.Minor = src.Spec.Upgrades.Minor

	dst.Spec.AddressSettings.ApplyRule = src.Spec.AddressSettings.ApplyRule
	dst.Spec.AddressSettings.AddressSetting = make([]AddressSettingType, len(src.Spec.AddressSettings.AddressSetting))
	for i, add := range src.Spec.AddressSettings.AddressSetting {
		dst.Spec.AddressSettings.AddressSetting[i] = AddressSettingType{
			DeadLetterAddress:                    add.DeadLetterAddress,
			AutoCreateDeadLetterResources:        add.AutoCreateDeadLetterResources,
			DeadLetterQueuePrefix:                add.DeadLetterQueuePrefix,
			DeadLetterQueueSuffix:                add.DeadLetterQueueSuffix,
			ExpiryAddress:                        add.ExpiryAddress,
			AutoCreateExpiryResources:            add.AutoCreateDeadLetterResources,
			ExpiryQueuePrefix:                    add.ExpiryQueuePrefix,
			ExpiryQueueSuffix:                    add.ExpiryQueueSuffix,
			ExpiryDelay:                          add.ExpiryDelay,
			MinExpiryDelay:                       add.MinExpiryDelay,
			MaxExpiryDelay:                       add.MaxExpiryDelay,
			RedeliveryDelay:                      add.RedeliveryDelay,
			MaxRedeliveryDelay:                   add.MaxRedeliveryDelay,
			MaxDeliveryAttempts:                  add.MaxDeliveryAttempts,
			MaxSizeBytes:                         add.MaxSizeBytes,
			MaxSizeBytesRejectThreshold:          add.MaxSizeBytesRejectThreshold,
			PageSizeBytes:                        add.PageSizeBytes,
			PageMaxCacheSize:                     add.PageMaxCacheSize,
			AddressFullPolicy:                    add.AddressFullPolicy,
			MessageCounterHistoryDayLimit:        add.MessageCounterHistoryDayLimit,
			LastValueQueue:                       add.LastValueQueue,
			DefaultLastValueQueue:                add.DefaultLastValueQueue,
			DefaultLastValueKey:                  add.DefaultLastValueKey,
			DefaultNonDestructive:                add.DefaultNonDestructive,
			DefaultExclusiveQueue:                add.DefaultExclusiveQueue,
			DefaultGroupRebalance:                add.DefaultGroupRebalance,
			DefaultGroupRebalancePauseDispatch:   add.DefaultGroupRebalancePauseDispatch,
			DefaultGroupBuckets:                  add.DefaultGroupBuckets,
			DefaultGroupFirstKey:                 add.DefaultGroupFirstKey,
			DefaultConsumersBeforeDispatch:       add.DefaultConsumersBeforeDispatch,
			DefaultDelayBeforeDispatch:           add.DefaultDelayBeforeDispatch,
			RedistributionDelay:                  add.RedistributionDelay,
			SendToDlaOnNoRoute:                   add.SendToDlaOnNoRoute,
			SlowConsumerThreshold:                add.SlowConsumerThreshold,
			SlowConsumerPolicy:                   add.SlowConsumerPolicy,
			SlowConsumerCheckPeriod:              add.SlowConsumerCheckPeriod,
			AutoCreateJmsQueues:                  add.AutoCreateJmsQueues,
			AutoDeleteJmsQueues:                  add.AutoDeleteJmsQueues,
			AutoCreateJmsTopics:                  add.AutoCreateJmsTopics,
			AutoDeleteJmsTopics:                  add.AutoDeleteJmsTopics,
			AutoCreateQueues:                     add.AutoCreateQueues,
			AutoDeleteQueues:                     add.AutoDeleteQueues,
			AutoDeleteCreatedQueues:              add.AutoDeleteCreatedQueues,
			AutoDeleteQueuesDelay:                add.AutoDeleteQueuesDelay,
			AutoDeleteQueuesMessageCount:         add.AutoDeleteQueuesMessageCount,
			ConfigDeleteQueues:                   add.ConfigDeleteQueues,
			AutoCreateAddresses:                  add.AutoCreateAddresses,
			AutoDeleteAddresses:                  add.AutoDeleteAddresses,
			AutoDeleteAddressesDelay:             add.AutoDeleteAddressesDelay,
			ConfigDeleteAddresses:                add.ConfigDeleteAddresses,
			ManagementBrowsePageSize:             add.ManagementBrowsePageSize,
			DefaultPurgeOnNoConsumers:            add.DefaultPurgeOnNoConsumers,
			DefaultMaxConsumers:                  add.DefaultMaxConsumers,
			DefaultQueueRoutingType:              add.DefaultQueueRoutingType,
			DefaultAddressRoutingType:            add.DefaultAddressRoutingType,
			DefaultConsumerWindowSize:            add.DefaultConsumerWindowSize,
			DefaultRingSize:                      add.DefaultRingSize,
			RetroactiveMessageCount:              add.RetroactiveMessageCount,
			EnableMetrics:                        add.EnableMetrics,
			Match:                                add.Match,
			ManagementMessageAttributeSizeLimit:  add.ManagementMessageAttributeSizeLimit,
			SlowConsumerThresholdMeasurementUnit: add.SlowConsumerThresholdMeasurementUnit,
			EnableIngressTimestamp:               add.EnableIngressTimestamp,
		}
	}

	log.Info("converted v1beta1 to v2alpha5")
	return nil
}
