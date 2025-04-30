import React, { useState } from "react";
import { baseModal } from "./Modal";
import {GetConfig, GetPaymentNetworkWalletAddr, SaveTunnelConfig} from "../../wailsjs/go/main/App";
import {main} from "../../wailsjs/go/models";
import Config = main.Config;

interface TunnelConfigurationModalProps {
    onClose: () => void;
    max: number;
    maxFree: number;
}

export const TunnelConfigurationModal: React.FC<TunnelConfigurationModalProps> = ({
                                                                                      onClose,
                                                                                      max,
                                                                                      maxFree,
                                                                                  }) => {
    const [nodes, setNodes] = useState<number>(1);
    const [addr, setAddr] = useState<string>("Loading...");
    const [enablePayments, setEnablePayments] = useState<boolean>(true);

    const handleIncrementNodes = () => setNodes((prev) => Math.min(prev + 1, 10));
    const handleDecrementNodes = () => setNodes((prev) => Math.max(prev - 1, 1));

    const isSaveDisabled = nodes > max;

    const handleContinue = () => {
        if (!isSaveDisabled) {
            SaveTunnelConfig(nodes, enablePayments).then(() => {
                onClose();
            });
        }
    };

    React.useEffect(() => {
        GetConfig().then((cfg: Config) => {
            setNodes(cfg.TunnelConfig!.TunnelSectionsNum);
            setEnablePayments(cfg.TunnelConfig!.PaymentsEnabled);
        });
        GetPaymentNetworkWalletAddr().then((addr: string) => {
            setAddr(addr);
        });
    }, []);

    const content = (
        <div className="new-tunnel-configuration">
            <h2 className="title">Tunnel Configuration</h2>

            <div className="config-block">
                <div className="field-group">
                    <span className="field-label">Number of Nodes</span>
                    <div className="nodes-control">
                        <button
                            onClick={handleDecrementNodes}
                            className="nodes-btn"
                            aria-label="Decrease nodes"
                        >
                            –
                        </button>
                        <span className="nodes-count">{nodes}</span>
                        <button
                            onClick={handleIncrementNodes}
                            className="nodes-btn"
                            aria-label="Increase nodes"
                        >
                            +
                        </button>
                    </div>
                </div>

                <div className="field-group checkbox-group">
                    <div className="checkbox-container">
                        <label className="payments-label">
                            <input
                                type="checkbox"
                                className="styled-checkbox"
                                checked={enablePayments}
                                onChange={(e) => setEnablePayments(e.target.checked)}
                            />
                            Enable Payments
                        </label>
                    </div>

                    {enablePayments && (
                        <>
                            <div className="ton-address">
                                <label htmlFor="tonAddress">TON Address:</label>
                                <input
                                    id="tonAddress"
                                    type="text"
                                    readOnly
                                    className="address-input"
                                    value={addr}
                                    onClick={(e) => e.currentTarget.select()}
                                />
                            </div>
                            <div className="important-text">
                                Please make sure there is enough TON for tunnel payments and payment-network contract deployment, deposit at least 5 TON.
                            </div>
                        </>
                    )}
                </div>
            </div>

            <div className="modal-control">
                <button
                    className="second-button"
                    onClick={() => {
                        onClose();
                    }}
                >
                    Cancel
                </button>
                <button
                    className="main-button"
                    onClick={handleContinue}
                    disabled={isSaveDisabled}
                >
                    Save
                </button>
            </div>
        </div>
    );

    return baseModal(onClose, content, 10);
};

export default TunnelConfigurationModal;