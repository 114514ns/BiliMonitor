import React, {useEffect} from 'react';
import {Modal, ModalBody, ModalContent, ModalFooter, ModalHeader} from "@heroui/react";
import Markdown from "react-markdown";
import axios from "axios";
import remarkGfm from 'remark-gfm'
import "github-markdown-css/github-markdown-light.css";
import rehypeRaw from "rehype-raw";

function NoticeDialog(props) {

    return (
        <div>
            <Modal onClose={props.onClose} isOpen={true} size={'xs'}>
                <ModalContent>
                    <ModalHeader>
                        About & Notice
                    </ModalHeader>
                    <ModalBody>
                        <div className={'markdown-body list-disc max-h-1/3 overflow-scroll' }>
                            <Markdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeRaw]}>{props.content}</Markdown>
                        </div>

                    </ModalBody>
                    <ModalFooter>

                    </ModalFooter>
                </ModalContent>
            </Modal>
        </div>
    );
}

export default NoticeDialog;