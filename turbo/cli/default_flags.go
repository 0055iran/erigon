// Copyright 2024 The Erigon Authors
// This file is part of Erigon.
//
// Erigon is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Erigon is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Erigon. If not, see <http://www.gnu.org/licenses/>.

package cli

import (
	"github.com/urfave/cli/v2"

	"github.com/erigontech/erigon/cmd/utils"
)

// DefaultFlags contains all flags that are used and supported by Erigon binary.
var DefaultFlags = []cli.Flag{
	&utils.DataDirFlag,
	&utils.EthashDatasetDirFlag,
	&utils.ExternalConsensusFlag,
	&utils.TxPoolDisableFlag,
	&utils.TxPoolLocalsFlag,
	&utils.TxPoolNoLocalsFlag,
	&utils.TxPoolPriceLimitFlag,
	&utils.TxPoolPriceBumpFlag,
	&utils.TxPoolBlobPriceBumpFlag,
	&utils.TxPoolAccountSlotsFlag,
	&utils.TxPoolBlobSlotsFlag,
	&utils.TxPoolTotalBlobPoolLimit,
	&utils.TxPoolGlobalSlotsFlag,
	&utils.TxPoolGlobalBaseFeeSlotsFlag,
	&utils.TxPoolAccountQueueFlag,
	&utils.TxPoolGlobalQueueFlag,
	&utils.TxPoolLifetimeFlag,
	&utils.TxPoolTraceSendersFlag,
	&utils.TxPoolCommitEveryFlag,
	&PruneDistanceFlag,
	&PruneBlocksDistanceFlag,
	&PruneModeFlag,
	&BatchSizeFlag,
	&BodyCacheLimitFlag,
	&DatabaseVerbosityFlag,
	&PrivateApiAddr,
	&PrivateApiRateLimit,
	&EtlBufferSizeFlag,
	&TLSFlag,
	&TLSCertFlag,
	&TLSKeyFlag,
	&TLSCACertFlag,
	&StateStreamDisableFlag,
	&SyncLoopThrottleFlag,
	&BadBlockFlag,

	&utils.HTTPEnabledFlag,
	&utils.HTTPServerEnabledFlag,
	&utils.GraphQLEnabledFlag,
	&utils.HTTPListenAddrFlag,
	&utils.HTTPPortFlag,
	&utils.AuthRpcAddr,
	&utils.AuthRpcPort,
	&utils.JWTSecretPath,
	&utils.HttpCompressionFlag,
	&utils.HTTPCORSDomainFlag,
	&utils.HTTPVirtualHostsFlag,
	&utils.AuthRpcVirtualHostsFlag,
	&utils.HTTPApiFlag,
	&utils.WSPortFlag,
	&utils.WSEnabledFlag,
	&utils.WsCompressionFlag,
	&utils.HTTPTraceFlag,
	&utils.HTTPDebugSingleFlag,
	&utils.StateCacheFlag,
	&utils.RpcBatchConcurrencyFlag,
	&utils.RpcStreamingDisableFlag,
	&utils.DBReadConcurrencyFlag,
	&utils.RpcAccessListFlag,
	&utils.RpcTraceCompatFlag,
	&utils.RpcGasCapFlag,
	&utils.RpcBatchLimit,
	&utils.RpcReturnDataLimit,
	&utils.AllowUnprotectedTxs,
	&utils.RpcMaxGetProofRewindBlockCount,
	&utils.RPCGlobalTxFeeCapFlag,
	&utils.TxpoolApiAddrFlag,
	&utils.TraceMaxtracesFlag,
	&HTTPReadTimeoutFlag,
	&HTTPWriteTimeoutFlag,
	&HTTPIdleTimeoutFlag,
	&AuthRpcReadTimeoutFlag,
	&AuthRpcWriteTimeoutFlag,
	&AuthRpcIdleTimeoutFlag,
	&EvmCallTimeoutFlag,
	&OverlayGetLogsFlag,
	&OverlayReplayBlockFlag,

	&RpcSubscriptionFiltersMaxLogsFlag,
	&RpcSubscriptionFiltersMaxHeadersFlag,
	&RpcSubscriptionFiltersMaxTxsFlag,
	&RpcSubscriptionFiltersMaxAddressesFlag,
	&RpcSubscriptionFiltersMaxTopicsFlag,

	&utils.SnapKeepBlocksFlag,
	&utils.SnapStopFlag,
	&utils.SnapStateStopFlag,
	&utils.DbPageSizeFlag,
	&utils.DbSizeLimitFlag,
	&utils.DbWriteMapFlag,
	&utils.TorrentPortFlag,
	&utils.TorrentMaxPeersFlag,
	&utils.TorrentConnsPerFileFlag,
	&utils.TorrentDownloadSlotsFlag,
	&utils.TorrentStaticPeersFlag,
	&utils.TorrentUploadRateFlag,
	&utils.TorrentDownloadRateFlag,
	&utils.TorrentVerbosityFlag,
	&utils.ListenPortFlag,
	&utils.P2pProtocolVersionFlag,
	&utils.P2pProtocolAllowedPorts,
	&utils.NATFlag,
	&utils.NoDiscoverFlag,
	&utils.DiscoveryV5Flag,
	&utils.NetrestrictFlag,
	&utils.NodeKeyFileFlag,
	&utils.NodeKeyHexFlag,
	&utils.DNSDiscoveryFlag,
	&utils.BootnodesFlag,
	&utils.StaticPeersFlag,
	&utils.TrustedPeersFlag,
	&utils.MaxPeersFlag,
	&utils.ChainFlag,
	&utils.DeveloperPeriodFlag,
	&utils.VMEnableDebugFlag,
	&utils.NetworkIdFlag,
	&utils.FakePoWFlag,
	&utils.GpoBlocksFlag,
	&utils.GpoPercentileFlag,
	&utils.InsecureUnlockAllowedFlag,
	&utils.IdentityFlag,
	&utils.CliqueSnapshotCheckpointIntervalFlag,
	&utils.CliqueSnapshotInmemorySnapshotsFlag,
	&utils.CliqueSnapshotInmemorySignaturesFlag,
	&utils.CliqueDataDirFlag,
	&utils.MiningEnabledFlag,
	&utils.ProposingDisableFlag,
	&utils.MinerNotifyFlag,
	&utils.MinerGasLimitFlag,
	&utils.MinerEtherbaseFlag,
	&utils.MinerExtraDataFlag,
	&utils.MinerNoVerfiyFlag,
	&utils.MinerSigningKeyFileFlag,
	&utils.MinerRecommitIntervalFlag,
	&utils.SentryAddrFlag,
	&utils.SentryLogPeerInfoFlag,
	&utils.DownloaderAddrFlag,
	&utils.DisableIPV4,
	&utils.DisableIPV6,
	&utils.NoDownloaderFlag,
	&utils.DownloaderVerifyFlag,
	&HealthCheckFlag,
	&utils.HeimdallURLFlag,
	&utils.WebSeedsFlag,
	&utils.WithoutHeimdallFlag,
	&utils.BorBlockPeriodFlag,
	&utils.BorBlockSizeFlag,
	&utils.WithHeimdallMilestones,
	&utils.WithHeimdallWaypoints,
	&utils.PolygonSyncFlag,
	&utils.PolygonSyncStageFlag,
	&utils.EthStatsURLFlag,
	&utils.OverridePragueFlag,

	&utils.CaplinDiscoveryAddrFlag,
	&utils.CaplinDiscoveryPortFlag,
	&utils.CaplinDiscoveryTCPPortFlag,
	&utils.CaplinCheckpointSyncUrlFlag,
	&utils.CaplinSubscribeAllTopicsFlag,
	&utils.CaplinMaxPeerCount,
	&utils.CaplinEnableUPNPlag,
	&utils.CaplinMaxInboundTrafficPerPeerFlag,
	&utils.CaplinMaxOutboundTrafficPerPeerFlag,
	&utils.CaplinAdaptableTrafficRequirementsFlag,
	&utils.SentinelAddrFlag,
	&utils.SentinelPortFlag,
	&utils.SentinelBootnodes,

	&utils.OtsSearchMaxCapFlag,

	&utils.SilkwormExecutionFlag,
	&utils.SilkwormRpcDaemonFlag,
	&utils.SilkwormSentryFlag,
	&utils.SilkwormVerbosityFlag,
	&utils.SilkwormNumContextsFlag,
	&utils.SilkwormRpcLogEnabledFlag,
	&utils.SilkwormRpcLogMaxFileSizeFlag,
	&utils.SilkwormRpcLogMaxFilesFlag,
	&utils.SilkwormRpcLogDumpResponseFlag,
	&utils.SilkwormRpcNumWorkersFlag,
	&utils.SilkwormRpcJsonCompatibilityFlag,

	&utils.BeaconAPIFlag,
	&utils.BeaconApiAddrFlag,
	&utils.BeaconApiAllowMethodsFlag,
	&utils.BeaconApiAllowOriginsFlag,
	&utils.BeaconApiAllowCredentialsFlag,
	&utils.BeaconApiPortFlag,
	&utils.BeaconApiReadTimeoutFlag,
	&utils.BeaconApiWriteTimeoutFlag,
	&utils.BeaconApiProtocolFlag,
	&utils.BeaconApiIdleTimeoutFlag,

	&utils.CaplinBackfillingFlag,
	&utils.CaplinBlobBackfillingFlag,
	&utils.CaplinDisableBlobPruningFlag,
	&utils.CaplinDisableCheckpointSyncFlag,
	&utils.CaplinArchiveFlag,
	&utils.CaplinEnableSnapshotGeneration,
	&utils.CaplinMevRelayUrl,
	&utils.CaplinValidatorMonitorFlag,
	&utils.CaplinCustomConfigFlag,
	&utils.CaplinCustomGenesisFlag,

	&utils.TrustedSetupFile,
	&utils.RPCSlowFlag,

	&utils.TxPoolGossipDisableFlag,
	&SyncLoopBlockLimitFlag,
	&SyncLoopBreakAfterFlag,
	&SyncParallelStateFlushing,

	&utils.ChaosMonkeyFlag,
	&utils.ExperimentalEFOptimizationFlag,
}
