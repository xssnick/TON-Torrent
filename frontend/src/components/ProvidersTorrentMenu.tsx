import React, {Component, useEffect, useState} from 'react';
import {EventsEmit, EventsOff, EventsOn, WindowSetMinSize} from "../../wailsjs/runtime";
import {
    BuildProviderContractData, BuildWithdrawalContractData,
    CheckHeader,
    GetFiles,
    GetPeers,
    GetProviderContract
} from "../../wailsjs/go/main/App";
import {
    TonConnectButton,
    TonConnectUI,
    useTonAddress,
    useTonConnectModal,
    useTonConnectUI,
    useTonWallet
} from "@tonconnect/ui-react";
import {textState} from "./Table";

export interface Provider {
    id: string
    lastProof: string
    proofEvery: string
    price: string
    pricePerProof: string
    status: string
    peer: string
    reason: string
    type: 'new' | 'committed' | 'removed'
    data: any
}

export interface ProvidersProps {
    torrent: string
}

interface State {
    address: string
    fetched: boolean
    providers: Provider[];
    contractBalance: string
}

function copy(text: string) {
    return (clicked: any) => {
        navigator.clipboard.writeText(text).then();
        clicked.target.classList.add('clicked');
        setTimeout(() => {
            clicked.target.classList.remove('clicked');
        },500);
    }
}

export const ProvidersTorrentMenu: React.FC<ProvidersProps> = (props) => {
    const [state, setState] = useState<State>({
        providers: [],
        fetched: false,
        address: '',
        contractBalance: '0 TON'
    });

    let address = useTonAddress(true);
    const [tonConnectUI] = useTonConnectUI();


    useEffect(() => {
        setState((current) => {
            return {...current, fetched: false, address: "", providers: [], contractBalance: "0 TON"}
        });

        let fetchContract = () => {
            if (!address) return;

            let initialTorrent = props.torrent;
            GetProviderContract(props.torrent, address).then(provider => {
                if (props.torrent != initialTorrent) return;

                if (!provider.Success) {
                    return
                }

                if (!provider.Deployed) {
                    setState((current) => {
                        // Load new providers from local cache
                        const cachedProviders = JSON.parse(localStorage.getItem('providers_' + props.torrent) || '[]') as Provider[];
                        for (const cachedProvider of cachedProviders) {
                            const existingProvider = current.providers.find(v => v.id === cachedProvider.id);
                            if (!existingProvider) {
                                current.providers.push({...cachedProvider, type: 'new'});
                            }
                        }

                        return {...current, providers: current.providers.filter(v => v.type == 'new'), fetched: true, address: ""}
                    });
                    return;
                }
                
                setState(current => {
                    const providersMap = new Map<string, Provider>();

                    // ① данные из контракта
                    if (provider.Providers) {
                        provider.Providers.forEach(p => {
                            providersMap.set(p.Key, {
                                id: p.Key,
                                lastProof: p.LastProof,
                                proofEvery: p.Span,
                                price: p.PricePerDay,
                                pricePerProof: p.PricePerProof,
                                status: p.Status,
                                reason: p.Reason,
                                peer: p.Peer,
                                type: 'committed',
                                data: p.Data,
                            });
                        });
                    }

                    current.providers
                        .filter(v => v.type === 'new' && !providersMap.has(v.id))
                        .forEach(v => providersMap.set(v.id, v));

                    const providers = Array.from(providersMap.values());

                    return {...current, providers: providers, fetched: true, contractBalance: provider.Balance, address: provider.Address}
                });
            });
        }

        EventsOn("provider-added", (p: any, hash: string) => {
            if (hash !== props.torrent) {
                return;
            }

            setState((current)=> {
                if (current.providers.find(v => v.id == p.Key)) {
                    return current;
                }

                let newProvider: Provider = {
                    id: p.Key,
                    lastProof: p.LastProof,
                    proofEvery: p.Span,
                    price: p.PricePerDay,
                    pricePerProof: p.PricePerProof,
                    status: "",
                    reason: "",
                    peer: "",
                    type: 'new',
                    data: p.Data,
                };

                current.providers.push(newProvider);

                // Save provider to local cache
                let cachedProviders = JSON.parse(localStorage.getItem('providers_' + props.torrent) || '[]') as Provider[];
                if (!cachedProviders.find(v => v.id === p.Key)) {
                    cachedProviders.push(newProvider);
                    localStorage.setItem('providers_'+props.torrent, JSON.stringify(cachedProviders));
                }

                return {...current, providers: current.providers}
            });
        });

        fetchContract();
        let inter = window.setInterval(fetchContract, 3000);

        tonConnectUI.onModalStateChange((s) => {
            if (s.status == 'closed') {
                WindowSetMinSize(800, 487);
            } else if (s.status == 'opened') {
                WindowSetMinSize(800, 720);
            }
        })

        return () => {
            EventsOff("provider-added");
            clearInterval(inter);
        };
    }, [props.torrent,address]);

    const statusSwitch = (status: string, peer: string) => {
        if (status === 'error') {
            return 'fail';
        } else if (status === 'downloading') {
            return 'downloading';
        } else if (status === 'active') {
            if (peer != '') {
                return 'seeding';
            }
            return 'active-op';
        }  else if (status === 'resolving' || status.startsWith('warning-')) {
            return 'searching';
        } else {
            return 'inactive';
        }
    }

    const renderProvidersList = () => {
        let items: any[] = [];

        for (const [i, t] of state.providers.entries()) {
            let cl = "";
            if (t.type == 'removed') {
                cl = "removed-provider"
            } else if (t.type == 'new') {
                cl = "new-provider";
            }

            let status = t.status;
            if (status.startsWith('warning-')) {
                status = status.slice(8)
            }

            let statusText = status.charAt(0).toUpperCase() + status.slice(1);
            let reason = t.reason;

            if ((status == 'balance' || status == 'active') && t.peer != '') {
                statusText = "Peer"
                reason += ", peer connected: "+ t.peer.slice(0,8);
            } else if (status == 'balance') {
                statusText = "Active"
            }

            items.push(
                <tr key={t.id} className={cl}>
                    <td style={{maxWidth:"200px"}}>{t.id}</td>
                    <td className={'small'} style={{display:"flex"}}>{ t.type == 'new' ? '' : <div className={"item-state-container"} onMouseEnter={(e) =>{
                        let tip = document.getElementById("tip");
                        tip!.textContent = reason != "" ? reason.charAt(0).toUpperCase() + reason.slice(1) : status.charAt(0).toUpperCase() + status.slice(1);
                        if (status == 'inactive') {
                            tip!.textContent = "Not connected"
                        }

                        let rectItem = document.getElementById("state-"+t.id)!.getBoundingClientRect()
                        let rectTip = tip!.getBoundingClientRect();

                        tip!.style.top =  (rectItem.y - (rectTip.height + 12)).toString()+"px";
                        tip!.style.left = (rectItem.x - (rectTip.width/2 - rectItem.width/2)).toString()+"px";

                        tip!.style.opacity = "1";
                        tip!.style.visibility = "visible";
                    }} onMouseLeave={
                        (e)=> {
                            let tip = document.getElementById("tip");
                            tip!.style.opacity = "0";
                            tip!.style.visibility = "hidden";
                        }
                    }><div id={"state-"+t.id} className={"item-state "+ statusSwitch(t.status, t.peer)}></div></div>}{status == 'inactive' ? "Proof "+t.lastProof : statusText}</td>
                    <td className={'small'}>{t.proofEvery}</td>
                    <td className={'small'}>{t.price}</td>
                    <td className={'small'}>{t.pricePerProof}</td>
                    <td className={'small'}><button icon-type="remove" onClick={()=>{
                        let prs = state.providers;

                        // Save provider to local cache
                        let cachedProviders = JSON.parse(localStorage.getItem('providers_' + props.torrent) || '[]') as Provider[];
                        if (cachedProviders.find(v => v.id === prs[i].id)) {
                            cachedProviders = cachedProviders.filter(v => v.id !== prs[i].id);
                            localStorage.setItem('providers_' + props.torrent, JSON.stringify(cachedProviders));
                        }

                        if (t.type == 'new') {
                            prs.splice(i,1);
                            localStorage.setItem('providers_'+props.torrent, JSON.stringify(prs));
                        } else if (t.type == 'removed') {
                            prs[i].type = 'committed';
                        } else if (t.type == 'committed') {
                            prs[i].type = 'removed';
                        }

                        setState((current)=> ({...current, providers: prs}));
                    }}/></td>
                </tr>
            );
        }
        return items;
    };

    const shortAddr = address.slice(0, 4) +"..."+ address.slice(-4)
    const shortContractAddr = state.address.length > 14 ? state.address.slice(0, 4) +"..."+ state.address.slice(-4) : "Not deployed"

    return (
            <>{address ? <div key={props.torrent} className="providers-connect">
                <div className="info-data-block" style={{ margin: "0 15px"}}>
                    <div className="basic" >
                        <div className="item" style={{width: "37%"}}><span className="field">Available balance: </span> { state.fetched ? <span className="value" style={{paddingLeft: "5px"}}>{state.contractBalance}</span> : <span className="loader" style={{height: "12px", width: "12px"}}/>}</div>
                        <div className="item" style={{width: "31%"}}><span className="field">Contract: </span>{ state.fetched ? <span className="value" style={{paddingLeft: "5px"}}>{shortContractAddr}</span> :  <span className="loader" style={{height: "12px", width: "12px"}}/>}{ state.address.length > 14 && state.fetched ? <button icon-type="copy" onClick={copy(state.address)}/> : <></>}</div>
                        <div className="item" style={{width: "32%"}}><span className="field">Authorized: </span><span className="value" style={{paddingLeft: "5px"}}>{shortAddr}</span><button icon-type="copy" onClick={copy(address)}/></div>
                        <div className="item" style={{flexGrow: "1", justifyContent: "flex-end"}}><button icon-type="logout" onClick={()=>{
                            tonConnectUI.disconnect().then();
                        }}/></div>
                    </div>
                </div>
                { !state.fetched ? <span className="loader" style={{height: "25px", width: "25px",margin: "5px auto 5px auto"}}/> : <></> }
                    { state.providers.length > 0 ? <table className="files-table">
                    <thead>
                        <tr>
                            <th>Provider key</th>
                            <th>State</th>
                            <th>Proof every</th>
                            <th>Price per day</th>
                            <th>Price per proof</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        {renderProvidersList()}
                    </tbody>
                </table> : "" }
                <div className="providers-login">
                    <button
                        className={state.providers.length > 0 ? "menu-item" : "menu-item main"}
                        onClick={() => {
                            EventsEmit("want_add_provider", props.torrent);
                        }}>
                        Add provider
                    </button>
                    <button
                        className={state.providers.length > 0 ? "menu-item main" : "menu-item"}
                        onClick={() => {
                            let providers = state.providers.filter(p => p.type != 'removed').map((p) => p.data);
                            let justTopup = state.address.length >= 14 && state.providers.find((x) => x.type != 'committed') == undefined;
                            EventsEmit("want_set_providers", props.torrent, address, providers, justTopup);
                        }}>
                        {(() => {
                            if (state.address.length < 14) {
                                return "Deploy contract";
                            }
                            if (state.providers.find((x) => x.type != 'committed')) {
                                return "Apply changes";
                            }
                            return "Topup balance";
                        })()}
                    </button>
                    <button
                        className="menu-item"
                        disabled={state.address.length < 14}
                        onClick={() => {
                            BuildWithdrawalContractData(props.torrent, address).then((tx)=>{
                                const transaction = {
                                    validUntil: Math.floor(Date.now() / 1000) + 90,
                                    messages: [
                                        {
                                            address: tx.Address,
                                            amount: tx.Amount,
                                            payload: tx.Body,
                                        }
                                    ]
                                }

                                tonConnectUI.sendTransaction(transaction, {
                                    modals: ['before', 'success', 'error'],
                                }).then((result)=> {
                                    console.log(result.boc);
                                });
                            })
                        }}>
                        Withdraw
                    </button>
                </div>
            </div> :

            <div className="providers-login"><button
                className="menu-item main"
                onClick={() => {tonConnectUI.openModal().then()}}>
                Connect Wallet
            </button></div>}</>
    );
};