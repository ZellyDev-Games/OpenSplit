import SessionPayload from "../../models/sessionPayload";

type SplitGameInfoProps = {
    sessionPayload: SessionPayload;
    completeClassName: string;
    showCompletedAttempts: boolean;
};

export default function SplitGameInfo({
    sessionPayload,
    completeClassName,
    showCompletedAttempts,
}: SplitGameInfoProps) {
    const splitFile = sessionPayload.loaded_split_file;

    return (
        <div id="gameInfo" className={completeClassName}>
            <h1 id="gameTitle" className={completeClassName}>
                <strong>{splitFile?.game_name}</strong>
            </h1>

            <h2 id="gameCategory" className={completeClassName}>
                <small>{splitFile?.game_category}</small>
            </h2>

            {splitFile?.variables
                .filter((variable) => variable.label?.trim())
                .map((variable) => (
                    <div key={variable.id} className="game-variable">
                        {variable.label}
                    </div>
                ))}

            <div id="attempts" className={completeClassName}>
                {showCompletedAttempts
                    ? `${splitFile?.runs.filter((run) => run.completed).length ?? 0}/${splitFile?.attempts ?? 0}`
                    : splitFile?.attempts}
            </div>
        </div>
    );
}
