import {DesktopOutlined} from "@ant-design/icons";

import { Card, CardHeader, CardBody, CardFooter, Avatar, Badge } from "@heroui/react";



function LiveCard({ liveData }) {
    const { Live, UName, UID, Area, Title } = liveData;
    /*
    return (
        <div>
            <Card
                key={UID}
                style={{ width: 300, marginRight: '20px' ,margin:'15px'}}
                actions={[
                    <Text strong type="secondary">Area: {Area}</Text>,
                    <Text strong type="secondary" onClick={() => {
                        window.open("https://space.bilibili.com/" + UID)
                    }}>UID: {UID}</Text>,
                ]}
            >
                <Badge
                    style={{ marginBottom: '10px' }}
                    status={Live ? 'success' : 'default'}
                    text={Live ? 'Live' : 'Offline'}
                />
                <Meta
                    avatar={<Avatar icon={<DesktopOutlined />} />}
                    title={UName}
                    description={Title}
                />
            </Card>

        </div>
    );

     */
    return (
        <div>
            <Card style={{ width: 300, marginRight: '20px' ,margin:'15px'}}>

                <CardHeader className="flex items-center gap-3">
                    <Badge color={Live?"success":"default"} content="">
                        <Avatar icon={<DesktopOutlined />} />
                    </Badge>

                    <div>
                        <h4 className="font-semibold">{UName}</h4>
                        <p className="text-gray-500">{Title}</p>
                    </div>
                </CardHeader>
                <CardBody>

                </CardBody>
                <CardFooter className="flex justify-between">
                    <span className="text-gray-500 font-semibold">Area: {Area}</span>
                    <span
                        className="text-blue-500 font-semibold cursor-pointer"
                        onClick={() => window.open(`https://space.bilibili.com/${UID}`)}
                    >
                    UID: {UID}
                </span>
                </CardFooter>
            </Card>
        </div>
    )
}

export default LiveCard;