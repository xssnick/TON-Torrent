import React, {useState} from "react";
import { Modal } from "./Modal";
import {main} from "../../wailsjs/go/models";
import SectionInfo = main.SectionInfo;
import {GetConfig, GetMaxTunnelNodes, GetPaymentNetworkWalletAddr} from "../../wailsjs/go/main/App";
import Config = main.Config;

interface TunnelNodesModalProps {
    sections: SectionInfo[];
    pricePerMBIn: string;
    pricePerMBOut: string;
    onCancel: () => void;
    onReroute: (num: number) => void;
    onAccept: () => void;
}

export const TunnelNodesModal: React.FC<TunnelNodesModalProps> = ({
                                                                      sections,
                                                                      pricePerMBIn,
                                                                      pricePerMBOut,
                                                                      onCancel,
                                                                      onReroute,
                                                                      onAccept,
                                                                  }) => {
    const [nodes, setNodes] = useState<number>(0);
    const [addr, setAddr] = useState<string>("Loading...");
    const [maxNodes, setMaxNodes] = useState<number>(0);

    const handleIncrementNodes = () => {
        let num = Math.min(nodes + 1, maxNodes);
        if (num != nodes)  {
            onReroute(num);
        }
        setNodes(num);
    };

    const handleDecrementNodes = () => {
        let num = Math.max(nodes - 1, 1);
        if (num != nodes)  {
            onReroute(num);
        }
        setNodes(num);
    };

    React.useEffect(() => {
        GetMaxTunnelNodes().then((num: number) => {
            setMaxNodes(num);
            GetConfig().then((cfg: Config) => {
                setNodes(Math.min(num, cfg.TunnelConfig!.TunnelSectionsNum));
            });
        });
        GetPaymentNetworkWalletAddr().then((addr: string) => {
            setAddr(addr);
        });
    }, []);

    const content = (
        <div className="tunnel-nodes-modal">
            <h2 className="modal-title">Tunnel Route</h2>

            <div className="config-block">
                <div className="field-group">
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
            </div>

            <div className="nodes-route">
                {sections.map((node, index) => (
                    <React.Fragment key={index}>
                        <div className="node-container">
                            {node.Outer && <div className="arrow up">↓</div>}
                            <span className={"node" + (node.Outer ? " outer" : "")}>{node.Name}</span>
                        </div>
                        {index !== sections.length - 1 && (
                            <span className="horizontal-arrow">{`↓`}</span>
                        )}
                    </React.Fragment>
                ))}
            </div>

            <div className="prices">
                <div className="price">
                    <span className="label">Price per MB (In):</span>
                    <span className="value">{pricePerMBIn} TON</span>
                </div>
                <div className="price">
                    <span className="label">Price per MB (Out):</span>
                    <span className="value">{pricePerMBOut} TON</span>
                </div>
            </div>

            {pricePerMBIn !== "0" || pricePerMBOut !== "0" ? <div className="config-block">
                <div className="field-group checkbox-group">
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
                        Please make sure there is enough TON for tunnel payments and payment-network contract deployment, deposit at least 5.5 TON.
                    </div>
                </div>
            </div>: null}

            <div className="modal-control">
                <button className="second-button" style={{ width: "100px" }} onClick={onCancel}>
                    Cancel
                </button>
                <button className="second-button" style={{ width: "100px" }} onClick={() => {
                    onReroute(nodes);
                }} disabled={nodes <= 0}>
                    Reroute
                </button>
                <button className="main-button" onClick={onAccept} disabled={nodes <= 0}>
                    Accept
                </button>
            </div>
        </div>
    );

    return <Modal allowClose={true} onHide={onCancel} content={content}/>;
};

export default TunnelNodesModal;