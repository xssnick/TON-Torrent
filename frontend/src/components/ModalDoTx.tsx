import React, {Component, useState} from 'react';
import {Modal} from "./Modal";
import {
    BuildProviderContractData,
    CancelCreateTorrent,
    CreateTorrent,
    ExportMeta,
    FetchProviderRates,
    OpenDir,
    OpenFolderSelectFile
} from "../../wailsjs/go/main/App";
import {useTonAddress, useTonConnectUI, useTonWallet} from "@tonconnect/ui-react";
import {ProvidersProps} from "./ProvidersTorrentMenu";

interface State {
    canContinue: boolean
    amount: string
}

interface DoTxModalProps {
    hash: string
    owner: string
    providers: any[]
    amount?: string
    justTopup: boolean
    onExit: () => void;
}

export const DoTxModal: React.FC<DoTxModalProps> = (props) => {
    const [state, setState] = useState<State>({
        canContinue: true,
        amount: props.amount ?? "0.5",
    });

    const [tonConnectUI] = useTonConnectUI();
    return <Modal allowClose={true} onHide={props.onExit} content={(
        <>
            <div style={{width: "287px"}} className="add-torrent-block">
                <span className="title">Contract deposit</span>
                <input className="torrent-name-input" placeholder={"Amount TON"} defaultValue={"0.5"} maxLength={20} autoFocus={true} onInput={(e) => {
                    let val = e.currentTarget.value;
                    let m = val.match(/([0-9]*[.])?[0-9]+/);
                    if (m && m[0] == val) {
                        let f = parseFloat(val);
                        if (f && f > 0.05) {
                            setState((current) => ({
                                ...current,
                                name: val,
                                canContinue: true,
                                amount: f.toFixed(6)
                            }));
                            return;
                        }
                    }
                    setState((current) => ({...current, name: val, canContinue: false}));
                }}/>
            </div>
            <div className="modal-control">
                <button className="second-button" onClick={props.onExit}>
                    Cancel
                </button>
                <button className="main-button" disabled={!state.canContinue} onClick={()=>{
                    BuildProviderContractData(props.hash, props.owner, state.amount, props.providers).then((tx)=>{
                        const transaction = {
                            validUntil: Math.floor(Date.now() / 1000) + 90,
                            messages: [
                                {
                                    address: tx.Address,
                                    amount: tx.Amount,
                                    stateInit: props.justTopup ? undefined : tx.StateInit,
                                    payload: props.justTopup ? undefined : tx.Body,
                                }
                            ]
                        }

                        tonConnectUI.sendTransaction(transaction, {
                            modals: ['before', 'success', 'error'],
                        }).then((result)=> {
                            console.log(result.boc);
                            props.onExit();
                        });
                    })
                }}>
                    Continue
                </button>
            </div>
        </>
    )}/>;
}