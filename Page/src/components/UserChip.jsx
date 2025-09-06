import React from 'react';
import {Avatar, Chip} from "@heroui/react";
import {CheckIcon} from "../pages/ChatPage";

function UserChip(props) {
    return (
        <div>
            <p className={'font-medium'}>{props.props.FromName}</p>
            {(
                <div className={'flex flex-row align-middle mt-2'}>
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
        </div>
    );
}

export default UserChip;