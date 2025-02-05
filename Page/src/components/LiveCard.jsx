import {DesktopOutlined} from "@ant-design/icons";
import {Avatar, Badge, Card, Space, Typography} from "antd";

const {Text} = Typography;
const {Meta} = Card;

function LiveCard({ liveData }) {
    const { Live, UName, UID, Area, Title } = liveData;
    return (
        <div>
            <Card
                key={UID}
                style={{ width: 300, marginRight: '20px' }}
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
}

export default LiveCard;