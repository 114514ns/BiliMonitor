import React, {useRef} from 'react';
import {Avatar, Chip, Dropdown, DropdownItem, DropdownMenu, DropdownTrigger} from "@heroui/react";
import {CheckIcon} from "../pages/ChatPage";
import { ContextMenu } from "radix-ui";

function UserChip(props) {
    const ref = React.useRef();
    return (
        <div className={'mt-2 w-full'}>
            <ContextMenu.Root>
                <ContextMenu.Trigger>
                    <p className={'font-medium w-full'} onContextMenu={(e) => {
                        ref.current.click()
                    }}>{props.props.FromName}</p>
                    {(
                        <div className={'flex flex-row align-middle'}>
                            <Avatar
                                src={`${protocol}://${host}:${port}${import.meta.env.PROD ? '' : '/api'}/face?mid=${props.props.FromId}`}
                                onClick={() => {
                                    toSpace(props.props.FromId);
                                }}/>

                            {props.props.MedalLevel ?              <Chip
                                startContent={props.props.MedalLevel != 0 ?<img src={getGuardIcon(props.props.GuardLevel)}/>:<CheckIcon size={18}/> }
                                variant="faded"
                                onClick={() => {
                                    toSpace(props.props.LiverID);
                                }}
                                style={{background: getColor(props.props.MedalLevel), color: 'white', marginLeft: '8px',marginTop:'4px'}}
                            >
                                {props.props.MedalName}
                                <span className="ml-2 text-xs font-bold px-2 py-0.5 rounded-full">
                                                            {props.props.MedalLevel}
                                                        </span>
                            </Chip>:<></>}
                        </div>
                    )}
                </ContextMenu.Trigger>
                <ContextMenu.Content>
                    <Dropdown>
                        <DropdownMenu aria-label="Static Actions">
                            <DropdownItem key="new">Bilibili</DropdownItem>
                            <DropdownItem key="copy">Copy link</DropdownItem>
                        </DropdownMenu>
                    </Dropdown>
                </ContextMenu.Content>
            </ContextMenu.Root>

        </div>
    );
}

export default UserChip;