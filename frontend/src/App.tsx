import React, {MouseEvent, Component} from 'react';
import LogoLight from "../public/light/logo.svg"
import ResizerLight from "../public/light/resizer.svg"
import DownloadLight from "../public/light/download.svg"
import LogoDark from "../public/dark/logo.svg"
import ResizerDark from "../public/dark/resizer.svg"
import DownloadDark from "../public/dark/download.svg"
import './tooltip.css';
import {Filter, Refresh, SelectedTorrent, Table} from "./components/Table";
import {AddTorrentModal} from "./components/ModalAddTorrent";
import {WaitReady, SetActive, WantRemoveTorrent, SwitchTheme, IsDarkTheme} from "../wailsjs/go/main/App";
import {FiltersMenu} from "./components/FiltersMenu";
import {BrowserOpenURL, EventsOn} from "../wailsjs/runtime";
import {FilesTorrentMenu} from "./components/FilesTorrentMenu";
import {CreateTorrentModal} from "./components/ModalCreateTorrent";
import {PeersTorrentMenu} from "./components/PeersTorrentMenu";
import {InfoTorrentMenu} from "./components/InfoTorrentMenu";
import {SettingsModal} from "./components/ModalSettings";
import {RemoveConfirmModal} from "./components/ModalRemoveConfirm";

interface State {
    isDark: boolean
    selectedItems:  SelectedTorrent[]
    infoSize: number
    tableFilter: Filter
    showAddTorrentModal: boolean
    showCreateTorrentModal: boolean
    showSettingsModal: boolean
    showRemoveConfirmModal: boolean

    overallUploadSpeed: string
    overallDownloadSpeed: string
    torrentMenuSelected: number

    ready: boolean

    openFileHash?: string
    removeHashes?: string[]
}

export class App extends Component<{}, State> {
    constructor(props: any, state: State) {
        super(props, state);

        this.state = {
            isDark: false,
            selectedItems: [],
            infoSize: 150,
            tableFilter: {
                type: "all",
                search: "",
            },
            showAddTorrentModal: false,
            showCreateTorrentModal: false,
            showSettingsModal: false,
            showRemoveConfirmModal: false,
            overallUploadSpeed: "0 Bytes",
            overallDownloadSpeed: "0 Bytes",
            torrentMenuSelected: -1,
            ready: false,
        }
    }

    setSelectedTorrentMenu = (n: number) => {
        return () => {
            this.setState((current)=>({...current, torrentMenuSelected: n}))
        }
    }

    hasActiveTorrents = () => {
        let has = false;
        this.state.selectedItems.forEach((si)=> {
            if (si.active) has = true;
        })
        return has;
    }
    hasInactiveTorrents = () => {
        let has = false;
        this.state.selectedItems.forEach((si)=> {
            if (!si.active) has = true;
        })
        return has;
    }

    toggleAddTorrentModal = () => {
        this.setState((current)=>({...current, showAddTorrentModal: !this.state.showAddTorrentModal, openFileHash: undefined}))
    }
    toggleCreateTorrentModal = () => {
        this.setState((current)=>({...current, showCreateTorrentModal: !this.state.showCreateTorrentModal}))
    }
    toggleSettingsModal = () => {
        this.setState((current)=>({...current, showSettingsModal: !this.state.showSettingsModal}))
    }
    toggleRemoveConfirmModal = () => {
        this.setState((current)=>({...current, showRemoveConfirmModal: !this.state.showRemoveConfirmModal, removeHashes: undefined}))
    }

    async componentDidMount() {
        let dark = await IsDarkTheme();
        this.setState((current)=>({...current, isDark: dark}))

        EventsOn("want_remove_torrent", (hashes: string[]) => {
            this.setState((current)=>({...current, removeHashes: hashes, showRemoveConfirmModal: true}))
        })
        EventsOn("open_torrent", (hash: string) => {
            this.setState((current)=>({...current, showAddTorrentModal: true, openFileHash: hash}))
        })
        EventsOn("daemon_ready", (data)=> {
            this.setState((current)=>({...current, ready: true}));
        })
        WaitReady().then()

        window.addEventListener('resize', () => {
            let inf = document.getElementsByClassName("torrent-info");
            if (inf.length == 0) {
                return
            }

            let topH = document.getElementsByClassName("top-bar")![0].getBoundingClientRect().height;
            let botH = document.getElementsByClassName("foot-bar")![0].getBoundingClientRect().height;
            let minINF = inf[0].getBoundingClientRect();

            let minH = minINF.height;
            if (minH > window.innerHeight-(topH+botH)) {
                this.setState((current)=>({...current, infoSize: window.innerHeight-(topH+botH)}));
            }
        })

        EventsOn("speed", (data)=> {
            this.setState((current)=>({...current, overallUploadSpeed: data.Upload, overallDownloadSpeed: data.Download}));
        })
    }

    extendInfoEvent = (mouseDownEvent: MouseEvent) =>  {
        const startSize = this.state.infoSize;
        const startPosition = { x: mouseDownEvent.pageX, y: mouseDownEvent.pageY };

        const onMouseMove = (mouseMoveEvent: any) => {
            let topH = document.getElementsByClassName("top-bar")![0].getBoundingClientRect().height+50;
            let botH = document.getElementsByClassName("foot-bar")![0].getBoundingClientRect().height;

            let sz = startSize + startPosition.y - mouseMoveEvent.pageY;
            if (sz > window.innerHeight-(topH+botH)) {
                sz = window.innerHeight-(topH+botH)
            }

            if (sz < 40) {
                sz = 40
            }
            this.setState((current)=>({...current, infoSize: sz}));
        }
        const onMouseUp = () => {
            document.body.removeEventListener("mousemove", onMouseMove);
        }
        document.body.addEventListener("mousemove", onMouseMove);
        document.body.addEventListener("mouseup", onMouseUp, { once: true });
    }

    render() {
        if (this.state.isDark) {
            document.documentElement.style.setProperty('--back', "#232328");
            document.documentElement.style.setProperty('--table-back', "#2D2D32");
            document.documentElement.style.setProperty('--card-border', "transparent");
            document.documentElement.style.setProperty('--text-primary', "#F3F3F6");
            document.documentElement.style.setProperty('--text-secondary', "#ACACAF");
            document.documentElement.style.setProperty('--torrent-menu-active', "rgba(255, 255, 255, 0.07)");
            document.documentElement.style.setProperty('--torrent-menu-inactive', "#36363C");
            document.documentElement.style.setProperty('--button-stroke', "#303035");
            document.documentElement.style.setProperty('--separator-alpha', "#4F4F53");
            document.documentElement.style.setProperty('--drop-border', "#303035");
            document.documentElement.style.setProperty('--drop-back', "transparent");
            document.documentElement.style.setProperty('--table-border', "#4F4F53");
            document.documentElement.style.setProperty('--button-back', "rgba(255, 255, 255, 0.07)");

            document.documentElement.style.setProperty("--search-img", "url(../dark/search.svg)");
            document.documentElement.style.setProperty("--close-img", "url(../dark/close.svg)");
            document.documentElement.style.setProperty("--play-img", "url(../dark/play.svg)");
            document.documentElement.style.setProperty("--pause-img", "url(../dark/pause.svg)");
            document.documentElement.style.setProperty("--close-disabled-img", "url(../dark/close-disabled.svg)");
            document.documentElement.style.setProperty("--play-disabled-img", "url(../dark/play-disabled.svg)");
            document.documentElement.style.setProperty("--pause-disabled-img", "url(../dark/pause-disabled.svg)");
            document.documentElement.style.setProperty("--settings-img", "url(../dark/settings.svg)");
            document.documentElement.style.setProperty("--theme-img", "url(../dark/theme.svg)");
            document.documentElement.style.setProperty("--copy-img", "url(../dark/copy.svg)");
            document.documentElement.style.setProperty("--expand-img", "url(../dark/expand.svg)");
        } else {
            document.documentElement.style.setProperty('--back', "#FFFFFF");
            document.documentElement.style.setProperty('--table-back', "#F7F9FB");
            document.documentElement.style.setProperty('--card-border', "#DDE3E6");
            document.documentElement.style.setProperty('--text-primary', "#04060B");
            document.documentElement.style.setProperty('--text-secondary', "#728A96");
            document.documentElement.style.setProperty('--torrent-menu-active', "rgba(118, 152, 187, 0.12)");
            document.documentElement.style.setProperty('--torrent-menu-inactive', "#EDF1F6");
            document.documentElement.style.setProperty('--button-stroke', "#E9EEF1");
            document.documentElement.style.setProperty('--separator-alpha', "#DFE5E8");
            document.documentElement.style.setProperty('--drop-border', "#E9EEF1");
            document.documentElement.style.setProperty('--drop-back', "#EDF1F6");
            document.documentElement.style.setProperty('--table-border', "rgba(0, 0, 0, 0.16)");
            document.documentElement.style.setProperty('--button-back', "rgba(118, 152, 187, 0.12)");

            document.documentElement.style.setProperty("--search-img", "url(../light/search.svg)");
            document.documentElement.style.setProperty("--close-img", "url(../light/close.svg)");
            document.documentElement.style.setProperty("--play-img", "url(../light/play.svg)");
            document.documentElement.style.setProperty("--pause-img", "url(../light/pause.svg)");
            document.documentElement.style.setProperty("--close-disabled-img", "url(../light/close-disabled.svg)");
            document.documentElement.style.setProperty("--play-disabled-img", "url(../light/play-disabled.svg)");
            document.documentElement.style.setProperty("--pause-disabled-img", "url(../light/pause-disabled.svg)");
            document.documentElement.style.setProperty("--settings-img", "url(../light/settings.svg)");
            document.documentElement.style.setProperty("--theme-img", "url(../light/theme.svg)");
            document.documentElement.style.setProperty("--copy-img", "url(../light/copy.svg)");
            document.documentElement.style.setProperty("--expand-img", "url(../light/expand.svg)");
        }

        return (
            <div id="App">
                <div className="daemon-waiter" style={this.state.ready ? {display: "none"} : {}}>
                    <div className="loader-block">
                        <span className="loader"/>
                    </div>
                </div>
                {this.state.showAddTorrentModal ? <AddTorrentModal openHash={this.state.openFileHash} onExit={this.toggleAddTorrentModal} isDark={this.state.isDark}/> : null}
                {this.state.showCreateTorrentModal ? <CreateTorrentModal onExit={this.toggleCreateTorrentModal}/> : null}
                {this.state.showSettingsModal ? <SettingsModal onExit={this.toggleSettingsModal}/> : null}
                {this.state.showRemoveConfirmModal ? <RemoveConfirmModal hashes={this.state.removeHashes!}  onExit={this.toggleRemoveConfirmModal} isDark={this.state.isDark}/> : null}
                <div className="left-bar">
                    <div className="logo-block">
                        <img className="logo-img" src={this.state.isDark ? LogoDark : LogoLight} alt=""/>
                    </div>
                    <div className="menu-block">
                        <FiltersMenu onChanged={(v) => {
                            this.setState((current)=>({...current, tableFilter: {
                                    type: v,
                                    search: this.state.tableFilter.search
                                }}));
                            Refresh();
                        }}/>
                    </div>
                    <div className="actions-menu">
                        <button className="menu-item main" onClick={this.toggleAddTorrentModal}>
                            Add
                        </button>
                        <button className="menu-item secondary" onClick={this.toggleCreateTorrentModal}>
                            Create
                        </button>
                    </div>
                </div>
                <div className="right-screen">
                    <div className="top-bar">
                        <div className="top-buttons-container">
                            <button className={this.hasInactiveTorrents() ? "top-button start" : "top-button start disabled"} style={{marginLeft: 0}} disabled={!this.hasInactiveTorrents()} onClick={() => {
                                this.state.selectedItems.forEach((t) => {
                                    SetActive(t.hash, true).then(Refresh)
                                })
                            }}/>
                            <button className={this.hasActiveTorrents() ? "top-button stop" : "top-button stop disabled"} disabled={!this.hasActiveTorrents()} onClick={() => {
                                this.state.selectedItems.forEach((t) => {
                                    SetActive(t.hash, false).then(Refresh)
                                })
                            }}/>
                            <button className={this.state.selectedItems.length > 0 ? "top-button remove" : "top-button remove disabled"} style={{marginRight: 0}} disabled={this.state.selectedItems.length == 0} onClick={() => {
                                WantRemoveTorrent(this.state.selectedItems.map(s => s.hash)).then(Refresh);
                            }}/>
                        </div>
                        <div className={"top-right"}>
                            <input type="text" className="search-input" placeholder="Search..." onChange={(e) => {
                                this.setState((current)=>({...current, tableFilter: {
                                        type: this.state.tableFilter.type,
                                        search: e.target.value,
                                    }}));
                            }}/>
                            <div className="top-buttons-container-right">
                                <button className={"top-button settings"} style={{marginLeft: 0}} onClick={this.toggleSettingsModal}/>
                                <button className={"top-button theme"} style={{marginRight: 0}} onClick={() => {
                                    SwitchTheme().then();
                                    this.setState((current)=>({...current, isDark: !this.state.isDark}));
                                }}/>
                            </div>
                        </div>
                    </div>
                    <div className="torrents-table">
                        <Table filter={this.state.tableFilter} onSelect={(sl) => {
                            let menu = this.state.torrentMenuSelected;
                            if (menu == -1 && sl.length > 0) {
                                menu = 0;
                            } else if (sl.length == 0) {
                                menu = -1;
                            }

                            this.setState((current) => ({ ...current, selectedItems: sl, torrentMenuSelected: menu }))
                        }}/>
                    </div>
                    { this.state.selectedItems.length >0 ? <div className="torrent-info" style={{minHeight: this.state.infoSize + "px",maxHeight: this.state.infoSize + "px"}}>
                        <div className="torrent-menu">
                            <div className="buttons-block">
                                <button disabled={this.state.torrentMenuSelected == 0 || this.state.torrentMenuSelected == -1} onClick={this.setSelectedTorrentMenu(0)}>Info</button>
                                <button disabled={this.state.torrentMenuSelected == 1 || this.state.torrentMenuSelected == -1} onClick={this.setSelectedTorrentMenu(1)}>Files</button>
                                <button disabled={this.state.torrentMenuSelected == 2 || this.state.torrentMenuSelected == -1} onClick={this.setSelectedTorrentMenu(2)}>Peers</button>
                            </div>
                            <div onMouseDown={this.extendInfoEvent} className="size-scroller"></div>
                            <div className="buttons-block">
                                <img style={{cursor: "ns-resize"}} onMouseDown={this.extendInfoEvent} src={this.state.isDark ? ResizerDark : ResizerLight}/>
                            </div>
                        </div>
                        <div className="torrent-body">
                            {this.state.torrentMenuSelected == 0 ? <InfoTorrentMenu torrent={this.state.selectedItems[0].hash}/> : ""}
                            {this.state.torrentMenuSelected == 1 ? <FilesTorrentMenu torrent={this.state.selectedItems[0].hash}/> : ""}
                            {this.state.torrentMenuSelected == 2 ? <PeersTorrentMenu torrent={this.state.selectedItems[0].hash}/> : ""}
                        </div>
                    </div> : ""}
                    <div className="foot-bar">
                        <div className="speed">
                            <span><img src={this.state.isDark ? DownloadDark : DownloadLight} alt=""/>{this.state.overallDownloadSpeed}</span>
                            <span><img className="upload" src={this.state.isDark ? DownloadDark : DownloadLight} alt=""/>{this.state.overallUploadSpeed}</span>
                        </div>
                    </div>
                </div>
            </div>
        )
    }
}

export default App
