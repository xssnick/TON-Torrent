import React, {Component} from 'react';
import {GetFiles, GetPlainFiles, GetTorrents, OpenFolder, OpenFolderSelectFile} from "../../wailsjs/go/main/App";
import {EventsEmit, EventsOff, EventsOn} from "../../wailsjs/runtime"; // let's also import Component

export interface FileItem {
    path: string
    name: string
    size: string
    downloaded: string
    progress: number
}

export interface FilesProps {
    torrent: string
}

interface State {
    files: FileItem[]
}

export class FilesTorrentMenu extends Component<FilesProps,State> {
    constructor(props: FilesProps, state:State) {
        super(props, state);

        this.state = {
            files: [],
        }
    }

    update() {
        GetPlainFiles(this.props.torrent).then((tr: any)=>{
            let newList: FileItem[] = []
            tr.forEach((t: any)=> {
                newList.push({
                    path: t.Path,
                    name: t.Name,
                    size: t.Size,
                    downloaded: t.Downloaded,
                    progress: t.Progress,
                })
            })

            this.setState({
                files: newList
            });
        });
    }

    componentDidMount() {
        this.update();
        EventsOn("update_files", () => {
            this.update();
        })
    }
    componentWillUnmount() {
        EventsOff("update_files")
    }

    renderFilesList() {
        let items = [];

        for (let t of this.state.files) {
            items.push(<tr onDoubleClick={() => {OpenFolderSelectFile(t.path).then()}}>
                <td style={{maxWidth:"150px"}}>{t.name}</td>
                <td className={"small"}>{t.size}</td>
                <td className={"small"}>{t.downloaded}</td>
                <td><div className="progress-block-small">
                    <span style={{textAlign:"left", width:"27px"}}>{t.progress}%</span>
                    <div className="progress-bar-small-form">
                        <div className="progress-bar-small" style={{width: t.progress+"%"}}></div>
                    </div></div>
                </td>
            </tr>);
        }
        return items;
    }

    render() {
        return <table className="files-table" style={{fontSize:12}}>
            <thead>
            <tr>
                <th>Name</th>
                <th style={{width:"100px"}}>Size</th>
                <th style={{width:"120px"}}>Downloaded</th>
                <th style={{width:"150px"}}>Progress</th>
            </tr>
            </thead>
            <tbody>
                {this.renderFilesList()}
            </tbody>
        </table>
    }
}