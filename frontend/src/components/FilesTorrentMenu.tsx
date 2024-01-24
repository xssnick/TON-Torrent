import React, {Component} from 'react';
import { GetPlainFiles, OpenFolderSelectFile} from "../../wailsjs/go/main/App";
import { EventsOff, EventsOn} from "../../wailsjs/runtime";

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
    allShown: boolean
}

export class FilesTorrentMenu extends Component<FilesProps,State> {
    constructor(props: FilesProps, state:State) {
        super(props, state);

        this.state = {
            files: [],
            allShown: true
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
                files: newList,
                allShown: newList.length < 1000
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
                <td style={{maxWidth: "80px"}}>{t.name}</td>
                <td>{t.size}</td>
                <td>{t.downloaded}</td>
                <td><div className="progress-block-small">
                    <span style={{textAlign:"left", width:"40px"}}>{t.progress}%</span>
                    <div className="progress-bar-small-form">
                        <div className="progress-bar-small" style={{width: t.progress+"%"}}></div>
                    </div></div>
                </td>
            </tr>);
        }

        if (!this.state.allShown) {
            items.push(<tr>
                <td style={{maxWidth: "200px"}}>
                    <span style={{textAlign:"center"}}>Too many files to render, please see others in directory</span>
                </td>
            </tr>);
        }
        return items;
    }

    render() {
        return <table className="files-table" style={{fontSize:12}}>
            <thead>
            <tr>
                <th style={{maxWidth: "80px"}}>Name</th>
                <th style={{width:"100px"}}>Size</th>
                <th style={{width:"120px"}}>Downloaded</th>
                <th style={{width:"130px"}}>Progress</th>
            </tr>
            </thead>
            <tbody>
                {this.renderFilesList()}
            </tbody>
        </table>
    }
}