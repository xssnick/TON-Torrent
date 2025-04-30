import React from "react";
import { baseModal } from "./Modal";
import {main} from "../../wailsjs/go/models";
import SectionInfo = main.SectionInfo;

interface TunnelNodesModalProps {
    sections: SectionInfo[];
    pricePerMBIn: string;
    pricePerMBOut: string;
    onCancel: () => void;
    onReroute: () => void;
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
    const content = (
        <div className="tunnel-nodes-modal">
            <h2 className="modal-title">Tunnel Route</h2>

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

            <div className="modal-control">
                <button className="second-button" style={{ width: "100px" }} onClick={onCancel}>
                    Cancel
                </button>
                <button className="second-button" style={{ width: "100px" }} onClick={onReroute}>
                    Reroute
                </button>
                <button className="main-button" onClick={onAccept}>
                    Accept
                </button>
            </div>
        </div>
    );

    return baseModal(onCancel, content);
};

export default TunnelNodesModal;