import React, {ReactElement, useEffect} from 'react';
import ReactDOM from 'react-dom';


interface ModalProps {
    content: ReactElement;
    onHide: () => void;
    allowClose?: boolean
}

export const Modal: React.FC<ModalProps> = ({
                                                content,
                                                onHide,
                                                allowClose,
                                            }) => {
    useEffect(() => {
        if (allowClose) {
            const handleKeyDown = (event: KeyboardEvent) => {
                if (event.key === "Escape") {
                    onHide();
                }
            };

            window.addEventListener("keydown", handleKeyDown);

            return () => {
                window.removeEventListener("keydown", handleKeyDown);
            };
        }
    }, []);

    const modal = (
        <>
            <div className="modal-backdrop" onClick={allowClose ? onHide : undefined}/>
            <div className="modal-wrap">
                <div className="modal-back">
                    <div className="modal-content">{content}</div>
                </div>
            </div>
        </>
    );
    return ReactDOM.createPortal(modal, document.body);
}