import React, {useState} from 'react';
import ReactDOM from 'react-dom';


export const useModal = () => {
    const [isShown, setIsShown] = useState<boolean>(false);
    const toggle = () => setIsShown(!isShown);
    return {
        isShown,
        toggle,
    };
};

export const baseModal = (onHide: () => void, content: JSX.Element) => {
    const modal = (
        <>
            <div className="modal-backdrop" onClick={onHide}/>
            <div className="modal-wrap">
                <div className="modal-back">
                    <div className="modal-content">{content}</div>
                </div>
            </div>
        </>
    );
    return ReactDOM.createPortal(modal, document.body);
}